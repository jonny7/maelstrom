package application

import (
	"os"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
	"github.com/rs/zerolog"
)

type App struct {
	Logger zerolog.Logger
	Agent  agent.Service
}

func New(consumer commander.Consumer, processor commander.Processor, client commander.HTTPClient, metrics metrics.Metrics) App {
	logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Str("service", "maelstrom application").Timestamp().Logger()
	a, err := agent.New(logger, consumer, processor, client, metrics)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	return App{
		Logger: logger,
		Agent:  a,
	}
}
