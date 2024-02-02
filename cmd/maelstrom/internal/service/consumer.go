package service

import (
	"fmt"

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/kafka"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

type consumer string

const (
	Kafka consumer = "kafka"
	Nop   consumer = "nop"
)

func NewConsumer(t consumer, logger logging.Logger, metrics metrics.Metrics) (commander.Consumer, error) {
	switch t {
	case Kafka:
		return kafka.MustNewKafka(logger, metrics)
	case Nop:
		return kafka.NewFake(), nil
	default:
		return nil, fmt.Errorf("unrecognized consumer type provider, expected either kafka or nop, but receieved: %s", t)
	}
}
