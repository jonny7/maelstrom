package service

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/consumer"
)

func NewApplication(consumer consumer.Consumer) application.App {
	return application.New(consumer)
}
