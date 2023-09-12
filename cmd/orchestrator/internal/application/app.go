package application

import (
	"github.com/rs/zerolog"
	"os"
)

type App struct {
	Logger zerolog.Logger
}

func New() App {
	return App{
		Logger: zerolog.New(os.Stdout).With().Str("service", "maelstrom gateway").Timestamp().Logger(),
	}
}
