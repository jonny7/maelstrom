package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/metrics"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/processor"
	a "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/api"
	u "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger, err := logging.NewLogger(logging.InfoLevel, 1, os.Stdout, "service", "Maelstrom Application")
	if err != nil {
		log.Fatal(err)
	}

	m := metrics.NewMetrics()

	go func() {
		// @todo move this
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":2112", nil)
	}()

	// @todo make env or arg
	client, err := service.NewConsumer(service.Kafka, logger, m)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	process := processor.NewProcessor(logger, m)
	if err != nil {
		log.Fatal(err)
	}

	// @todo same with processor
	app := service.NewApplication(client, process, &http.Client{}, logger, m)

	api := a.New(app)

	errs := make(chan error)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})

	api.Run(done, errs, func(router chi.Router) http.Handler {
		return maelstrom.HandlerFromMux(api, router)
	})

	// @todo headless
	ui := u.New(app)
	ui.Run(done, errs, func(router chi.Router) http.Handler {
		return maelstrom.HandlerFromMux(ui, router)
	})

	for {
		select {
		case msg := <-sig:
			logger.Log(logging.InfoLevel, fmt.Sprintf("shutting down from signal: %v", msg))
			close(done)
			return
		case err = <-errs:
			logger.LogWithError(logging.ErrorLevel, "returning from ListenAndServe", err)
			return
		}
	}
}
