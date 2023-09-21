package application

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/adapters/driven/subscriber/inmem"
	"github.com/rs/zerolog"
	"os"
)

type App struct {
	Logger     zerolog.Logger
	Subscriber *inmem.InMemorySubscriber
}

func New() App {
	logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom application").Timestamp().Logger()

	workerRepository := inmem.NewInMemoryRepository()

	return App{
		Logger:     logger,
		Subscriber: workerRepository,
	}
}
