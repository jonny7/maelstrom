package consumer

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/consumer"
)

type Service struct {
	consumer.Consumer
}

func NewConsumer(c consumer.Consumer) Service {
	return Service{c}
}
