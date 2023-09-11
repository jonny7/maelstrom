package application

import (
	"github.com/rs/zerolog"
	"os"
)

type Application struct {
	Logger *zerolog.Logger
}

func New() Application {
	logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom gateway").Timestamp().Logger()

	return Application{
		Logger: &logger,
	}

}

func (a Application) HealthHandler() bool {
	return true
}
