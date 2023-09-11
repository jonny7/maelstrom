package main

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/adapters/primary/http"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application/config"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application/services"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	app := services.NewApplication()
	server := http.New(app, *cfg)
	if se := server.Run(); err != nil {
		log.Error().Err(se).Send()
	}
}
