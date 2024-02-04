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
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/kafka"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/metrics/prometheus"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/process"
	a "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/api"
	u "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/consumer"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/metrics"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/processor"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/sender"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger, err := logging.NewLogger(logging.InfoLevel, 1, os.Stdout, "service", "Maelstrom Application")
	if err != nil {
		log.Fatal(err)
	}

	// @todo add flag to disable/enable
	m := metrics.NewMetrics(prometheus.NewMetrics())

	go func() {
		// @todo move this
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":2112", nil)
	}()

	// @todo make env or arg
	c, err := kafka.NewFake(m) //kafka.MustNewKafka(logger, m)
	if err != nil {
		log.Fatal(err)
	}

	client := consumer.NewConsumer(c)
	defer client.Close()

	proc := process.NewProcessor(logger, m)
	if err != nil {
		log.Fatal(err)
	}

	p := processor.NewProcessor(proc)

	h := sender.NewSender(&http.Client{})

	app := service.NewApplication(client, p, h, logger, m)

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
