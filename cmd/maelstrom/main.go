package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	h "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/http"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service"
	"github.com/rs/zerolog/log"
)

func main() {
	app := service.NewApplication()

	api := h.New(app)

	errs := make(chan error)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	api.Run(errs, func(router chi.Router) http.Handler {
		return maelstrom.HandlerFromMux(api, chi.NewRouter())
	})

	// @todo headless
	u := ui.New(app)

	u.Run(errs)

	for {
		select {
		case <-sig:
			log.Info().Msgf("shutting down from signal: %v", sig)
		case err := <-errs:
			log.Err(err).Msgf("returning from ListenAndServe: %v", err)
		}
	}
}
