package main

import (
	_ "embed"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/kafka"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/processor"
	a "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/api"
	u "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service"
	"github.com/rs/zerolog/log"
)

func main() {
	// @todo move to dynamic config approach
	client, err := kafka.New()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer client.Close()

	// @todo same with processor
	app := service.NewApplication(client, processor.RequestProcessor{})

	api := a.New(app)

	errs := make(chan error)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	api.Run(errs, func(router chi.Router) http.Handler {
		return maelstrom.HandlerFromMux(api, router)
	})

	// @todo headless
	ui := u.New(app)
	ui.Run(errs, func(router chi.Router) http.Handler {
		return maelstrom.HandlerFromMux(ui, router)
	})

	for {
		select {
		case <-sig:
			log.Info().Msgf("shutting down from signal: %v", sig)
		case err = <-errs:
			log.Err(err).Msgf("returning from ListenAndServe: %v", err)
		}
	}
}
