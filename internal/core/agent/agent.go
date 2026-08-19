package agent

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/jonny7/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/internal/core/metrics"
	"github.com/jonny7/maelstrom/internal/logging"
)

type agent struct {
	// commander runs and tracks load-test vortexes across the Maelstrom nodes
	commander commander.Command
	// membership is this node's view of the cluster, provided by a driven adapter
	membership Membership
	logger     logging.Logger
}

//go:generate mockgen -source=agent.go -destination mocks/agent.go -package mocks

type Service interface {
	Members() []Member
	Start(host string, jobs, workers int, cbuf, rbuf int)
	Stop(id uuid.UUID)
	Vortexes() []commander.Vortex
	Leave() error
}

func (a *agent) Start(host string, jobs, workers int, cbuf, rbuf int) {
	a.logger.Log(logging.DebugLevel, "starting generator and processor")
	a.commander.Start(host, jobs, workers, cbuf, rbuf)
}

func (a *agent) Stop(id uuid.UUID) {
	a.logger.Log(logging.DebugLevel, fmt.Sprintf("stop load test %v was triggered by the user", id))
	a.commander.Stop(id)
}

func (a *agent) Members() []Member {
	return a.membership.Members()
}

// Leave announces this node's departure to the cluster so peers mark it
// left immediately instead of failed.
func (a *agent) Leave() error {
	return a.membership.Leave()
}

func (a *agent) Vortexes() []commander.Vortex {
	return a.commander.Vortexes()
}

// New wires the commander and the injected membership into an agent.
func New(logger logging.Logger, generator commander.Generator, processor commander.Processor, client commander.HTTPDoer, metrics metrics.Metrics, membership Membership) Service {
	return &agent{
		commander:  commander.NewCommander(generator, processor, client, logger, metrics),
		membership: membership,
		logger:     logger,
	}
}
