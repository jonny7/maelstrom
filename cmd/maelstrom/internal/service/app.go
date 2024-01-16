package service

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
)

func NewApplication(consumer commander.Consumer, processor commander.Processor, client commander.HTTPClient) application.App {
	return application.New(consumer, processor, client)
}
