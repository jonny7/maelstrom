package application

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/domain/worker"
	"github.com/rs/zerolog"
	"os"
)

type App struct {
	Logger     zerolog.Logger
	Subscriber worker.Service
}

func New(store worker.Store) App {
	logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom application").Timestamp().Logger()

	return App{
		Logger:     logger,
		Subscriber: worker.NewService(store),
	}
}
