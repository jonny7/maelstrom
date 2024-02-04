package application

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/consumer"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

type App struct {
	Logger logging.Logger
	Agent  agent.Service
}

func New(consumer consumer.Consumer, processor processor.Processor, client sender.HTTPDoer, logger logging.Logger, metrics metrics.Metrics) App {
	a, err := agent.New(logger, consumer, processor, client, metrics)
	if err != nil {
		logger.LogWithError(logging.ErrorLevel, "unable to initialize agent", err)
	}

	return App{
		Logger: logger,
		Agent:  a,
	}
}
