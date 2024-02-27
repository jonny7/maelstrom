package service

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/generator"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

func NewApplication(generator generator.Generator, processor processor.Processor, client sender.HTTPDoer, logger logging.Logger, metrics metrics.Metrics) application.App {
	return application.New(generator, processor, client, logger, metrics)
}
