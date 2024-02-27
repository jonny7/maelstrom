package application

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/generator"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/k8s"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
	svc "github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/agent"
)

type App struct {
	Logger logging.Logger
	Agent  svc.Service
	K8s    k8s.K8s
}

func New(generator generator.Generator, processor processor.Processor, client sender.HTTPDoer, logger logging.Logger, metrics metrics.Metrics) App {
	a, err := agent.New(logger, generator, processor, client, metrics)
	if err != nil {
		logger.LogWithError(logging.ErrorLevel, "unable to initialize agent", err)
	}

	// @todo add disable flag k8s for running locally
	// do better org here
	k, e := k8s.New("default")
	if e != nil {
		logger.LogWithError(logging.ErrorLevel, "failed to initialize k8s", e)
	}

	return App{
		Logger: logger,
		Agent:  svc.NewAgentService(a),
		K8s:    k, //k8s.NewFakeK8s(),
	}
}
