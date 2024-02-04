package service

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/consumer"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/metrics"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/processor"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/sender"
)

func NewApplication(consumer consumer.Service, processor processor.Service, client sender.Service, logger logging.Logger, metrics metrics.Service) application.App {
	return application.New(consumer, processor, client, logger, metrics)
}
