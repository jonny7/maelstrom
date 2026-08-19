// Package application is the application layer
package application

import (
	"errors"

	"github.com/google/uuid"

	"github.com/jonny7/maelstrom/internal/application/dto"
	"github.com/jonny7/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/internal/core/metrics"
	"github.com/jonny7/maelstrom/internal/logging"
)

// Scaler is the driven port for resizing the cluster of Maelstrom nodes.
type Scaler interface {
	Scale(replicas int) error
}

// ErrInvalidReplicas rejects a Scale request for a negative replica count.
var ErrInvalidReplicas = errors.New("replicas must be zero or greater")

type App struct {
	agent  agent.Service
	scaler Scaler
}

// New wires the core services into an App. All driven ports (generator,
// processor, client, metrics, scaler, membership) are constructed by the caller.
func New(generator commander.Generator, processor commander.Processor, client commander.HTTPDoer, logger logging.Logger, metrics metrics.Metrics, scaler Scaler, membership agent.Membership) App {
	return App{
		agent:  agent.New(logger, generator, processor, client, metrics, membership),
		scaler: scaler,
	}
}

// Membership returns the current cluster members.
func (a App) Membership() []dto.Member {
	return dto.MemberDTO(a.agent.Members())
}

// Leave gracefully removes this node from the cluster; call on shutdown.
func (a App) Leave() error {
	return a.agent.Leave()
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
func (a App) Scale(replicas int) error {
	if replicas < 0 {
		return ErrInvalidReplicas
	}
	return a.scaler.Scale(replicas)
}
