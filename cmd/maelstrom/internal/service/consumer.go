package service

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/kafka"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
)

type consumer string

const (
	Kafka consumer = "kafka"
)

func NewConsumer(t consumer, logger logging.Logger) (commander.Consumer, error) {
	switch t {
	case Kafka:
		return kafka.MustNewKafka(logger)
	default:
		return kafka.NewFake(), nil
	}
}
