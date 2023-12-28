package application

import (
	"os"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/agent"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/commander"
	"github.com/rs/zerolog"
)

type App struct {
	Logger zerolog.Logger
	Agent  *agent.Agent
}

func New(consumer commander.Consumer) App {
	logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom application").Timestamp().Logger()
	a, err := agent.New(logger, consumer)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	return App{
		Logger: logger,
		Agent:  a,
	}
}
