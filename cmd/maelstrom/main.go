package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	h "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/http"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/http/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	app := service.NewApplication()

	server := h.New(app, *cfg)

	server.Run(func(router chi.Router) http.Handler {
		return maelstrom.HandlerFromMux(server, chi.NewRouter())
	})
}
