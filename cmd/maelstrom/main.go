package main

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/http"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/http/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	app := service.NewApplication()

	server := http.New(app, *cfg)

	if se := server.Run(); se != nil {
		log.Error().Err(se).Send()
	}
}
