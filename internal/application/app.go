// Package application is the application layer
package application

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/jonny7/maelstrom/internal/application/dto"
	"github.com/jonny7/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/internal/core/metrics"
	"github.com/jonny7/maelstrom/internal/logging"
)

// Cluster is the driven port for managing the cluster of Maelstrom nodes.
type Cluster interface {
	Scale(replicas int) error
	DeleteNode(node string) error
}

// ErrInvalidReplicas rejects a Scale request for a negative replica count.
var ErrInvalidReplicas = errors.New("replicas must be zero or greater")

// ErrMissingNode rejects a DeleteNode request without a node name.
var ErrMissingNode = errors.New("node name is required")

type App struct {
	agent   agent.Service
	cluster Cluster
}

// New wires the core services into an App. Configuration and all driven ports
// (generator, processor, client, metrics, scaler) are constructed by the caller.
func New(cfg agent.Config, generator commander.Generator, processor commander.Processor, client commander.HTTPDoer, logger logging.Logger, metrics metrics.Metrics, cluster Cluster) (App, error) {
	a, err := agent.New(cfg, logger, generator, processor, client, metrics)
	if err != nil {
		return App{}, fmt.Errorf("unable to initialize agent: %w", err)
	}

	return App{
		agent:   a,
		cluster: cluster,
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
func (a App) Scale(replicas int) error {
	if replicas < 0 {
		return ErrInvalidReplicas
	}
	return a.cluster.Scale(replicas)
}

// DeleteNode kills the named node; in k8s the StatefulSet replaces it.
func (a App) DeleteNode(name string) error {
	if name == "" {
		return ErrMissingNode
	}
	return a.cluster.DeleteNode(name)
}
