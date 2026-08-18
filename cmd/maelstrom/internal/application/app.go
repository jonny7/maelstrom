// Package application is the application layer
package application

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application/dto"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/generator"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/k8s"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

type App struct {
	agent agent.Service
	k8s   k8s.K8s
}

// New wires the core services into an App. All driven ports (generator, processor,
// client, metrics, scaler) are constructed by the caller.
func New(generator generator.Generator, processor processor.Processor, client sender.HTTPDoer, logger logging.Logger, metrics metrics.Metrics, scaler k8s.K8s) (App, error) {
	a, err := agent.New(logger, generator, processor, client, metrics)
	if err != nil {
		return App{}, fmt.Errorf("unable to initialize agent: %w", err)
	}

	return App{
		agent: a,
		k8s:   scaler,
	}, nil
}

// Membership returns the current cluster members.
func (a App) Membership() []dto.Member {
	return dto.MemberDTO(a.agent.Members())
}

// StartVortex begins a new load-test run.
func (a App) StartVortex(host string, jobs, workers, cbuf, rbuf int) {
	a.agent.Start(host, jobs, workers, cbuf, rbuf)
}

// EndVortex stops the load-test run with the given id.
func (a App) EndVortex(id uuid.UUID) {
	a.agent.Stop(id)
}

// Vortexes returns all tracked load-test runs.
func (a App) Vortexes() []commander.Vortex {
	return a.agent.Vortexes()
}

// Scale resizes the Maelstrom cluster to the given replica count.
func (a App) Scale(replicas int) (int, error) {
	return a.k8s.Scale(replicas)
}
