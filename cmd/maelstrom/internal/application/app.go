package application

import (
	"os"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/agent"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/consumer"
	"github.com/rs/zerolog"
)

type App struct {
	Logger   zerolog.Logger
	Agent    *agent.Agent
	Consumer consumer.Consumer
}

func New(consumer consumer.Consumer) App {
	logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom application").Timestamp().Logger()
	a, err := agent.New(logger)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	return App{
		Logger:   logger,
		Agent:    a,
		Consumer: consumer,
	}
}
