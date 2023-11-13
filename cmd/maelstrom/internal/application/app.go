package application

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/agent"
	"github.com/rs/zerolog"
	"os"
)

type App struct {
	Logger zerolog.Logger
	Agent  *agent.Agent
}

func New() App {
	logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom application").Timestamp().Logger()
	a, err := agent.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	return App{
		Logger: logger,
		Agent:  a,
	}
}
