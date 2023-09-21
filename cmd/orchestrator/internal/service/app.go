package service

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application"
)

func NewApplication() application.App {
	return application.New()
}
