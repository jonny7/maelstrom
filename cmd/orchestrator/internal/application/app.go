package application

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application/domain/worker"
	"github.com/rs/zerolog"
	"os"
)

type App struct {
	Logger     zerolog.Logger
	Subscriber *worker.Service
}

func New() App {
	logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom application").Timestamp().Logger()
	service, err := worker.NewService(worker.InMemory)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	return App{
		Logger:     logger,
		Subscriber: service,
	}
}
