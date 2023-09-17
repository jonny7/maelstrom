package application

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application/domain/worker"
)

type App struct {
	//Logger     zerolog.Logger
	Subscriber worker.Subscriber
}

func New() App {
	//logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom gateway").Timestamp().Logger()

	subscribe, _ := worker.NewSubscriber(worker.InMemory)

	return App{
		//	Logger:     logger,
		Subscriber: subscribe,
	}
}
