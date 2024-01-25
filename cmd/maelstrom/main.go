package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/metrics"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/processor"
	a "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/api"
	u "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service"
)

func main() {
	// @todo make env or arg
	client, err := service.NewConsumer(service.Kafka)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// @todo same with processor
	app := service.NewApplication(client, processor.NewProcessor(), &http.Client{}, metrics.New())

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
			log.Printf("shutting down from signal: %v", msg)
			close(done)
			return
		case err = <-errs:
			log.Printf("returning from ListenAndServe: %v", err)
			return
		}
	}
}
