package agent

import (
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/soheilhy/cmux"

	"github.com/jonny7/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/internal/core/membership"
	"github.com/jonny7/maelstrom/internal/core/metrics"
	"github.com/jonny7/maelstrom/internal/logging"
)

type agent struct {
	// Config is the agents configuration
	config Config
	// commander runs and tracks load-test vortexes across the Maelstrom nodes
	commander commander.Command
	// mux helps serve UDP & TCP over the same port
	mux cmux.CMux
	// membership configures Serf and eventing
	membership *membership.Membership
	logger     logging.Logger
}

//go:generate mockgen -source=agent.go -destination mocks/agent.go -package mocks

type Service interface {
	Members() []membership.Member
	Start(host string, jobs, workers int, cbuf, rbuf int)
	Stop(id uuid.UUID)
	Vortexes() []commander.Vortex
}

func (a *agent) Start(host string, jobs, workers int, cbuf, rbuf int) {
	a.logger.Log(logging.DebugLevel, "starting generator and processor")
	a.commander.Start(host, jobs, workers, cbuf, rbuf)
}

func (a *agent) Stop(id uuid.UUID) {
	a.logger.Log(logging.DebugLevel, fmt.Sprintf("stop load test %v was triggered by the user", id))
	a.commander.Stop(id)
}

func (a *agent) Members() []membership.Member {
	return a.membership.Members()
}

func (a *agent) Vortexes() []commander.Vortex {
	return a.commander.Vortexes()
}

// New returns a new agent or errors. Configuration is provided by the caller.
// The agent will also set up all membership for the Serf cluster
func New(cfg Config, logger logging.Logger, generator commander.Generator, processor commander.Processor, client commander.HTTPDoer, metrics metrics.Metrics) (Service, error) {
	// create agent
	a := &agent{
		config: cfg,
		logger: logger,
	}

	// create commander
	a.commander = commander.NewCommander(generator, processor, client, logger, metrics)

	// setup mux or err
	if err := a.setupMux(); err != nil {
		return nil, err
	}
	// multiplex port
	go func() {
		err := a.serve()
		if err != nil {
			logger.LogWithError(logging.ErrorLevel, "unable to multiplex agent", err)
		}
	}()

	// setup membership or err
	if err := a.setupMembership(logger); err != nil {
		return nil, err
	}
	return a, nil
}

// setupMembership configures the membership for this Serf node or errors
func (a *agent) setupMembership(logger logging.Logger) error {
	rpcAddr, err := a.config.RPCAddr()
	if err != nil {
		return err
	}

	a.membership, err = membership.New(a.commander, membership.Config{
		NodeName: a.config.NodeName,
		BindAddr: a.config.BindAddr,
		Tags: map[string]string{
			"rpc_addr": rpcAddr,
		},
		StartJoinAddrs: a.config.StartJoinAddrs,
	}, logger)
	return err
}

// serve serves multiplexed TCP and UDP protocols across a single port
func (a *agent) serve() error {
	if err := a.mux.Serve(); err != nil {
		a.logger.LogWithError(logging.ErrorLevel, "serf is shutting down", err)
		return err
	}
	return nil
}

// setupMux creates a new connection multiplexer
func (a *agent) setupMux() error {
	addr, err := net.ResolveTCPAddr("tcp", a.config.BindAddr)
	if err != nil {
		return err
	}
	rpcAddr := fmt.Sprintf("%s:%d", addr.IP.String(), a.config.RPCPort)
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}
	a.mux = cmux.New(ln)
	return nil
}
