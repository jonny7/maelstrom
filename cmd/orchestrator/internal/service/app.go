package service

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/domain/worker"
)

func NewApplication(store worker.Store) application.App {
	return application.New(store)
}
