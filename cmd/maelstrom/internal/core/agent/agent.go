package agent

//go:generate mockgen -source agent.go -destination mocks/mock_agent.go

import (
	"fmt"
	"net"

	"github.com/caarlos0/env/v10"
	"github.com/hashicorp/serf/serf"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/membership"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soheilhy/cmux"
)

type agent struct {
	// Config is the agents configuration
	config Config
	// Commander is a raft based set of changes to apply across the Maelstrom nodes
	commander *commander.Commander
	// mux helps serve UDP & TCP over the same port
	mux cmux.CMux
	// membership configures Serf and eventing
	membership *membership.Membership
	// shutdowns receive channel events, signifying it should be shut down
	shutdowns chan struct{}
	//shutdownLock sync.Mutex @todo make this graceful
}

type Agent struct {
	agent *agent
}

type Service interface {
	Members() []serf.Member
}

func (a Agent) Members() []serf.Member {
	return a.agent.membership.Members()
}

// New returns a new agent or errors. The main configuration is provided through environment vars or defaults.
// The agent will also set up all membership for the Serf cluster
func New(logger zerolog.Logger, consumer commander.Consumer, processor commander.Processor) (Service, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	// create agent
	a := &agent{
		config:    cfg,
		shutdowns: make(chan struct{}),
	}

	// create commander
	cmdr, err := commander.NewCommander(cfg.Commander, consumer, processor)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create commander")
	}
	a.commander = cmdr

	// setup mux or err
	if err = a.setupMux(); err != nil {
		return nil, err
	}
	// multiplex port
	go func() {
		err = a.serve()
		if err != nil {
			logger.Error().Err(err).Send()
		}
	}()

	// setup membership or err
	if err := a.setupMembership(logger); err != nil {
		return nil, err
	}
	return &Agent{agent: a}, nil
}

// setupMembership configures the membership for this Serf node or errors
func (a *agent) setupMembership(logger zerolog.Logger) error {
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
		log.Error().Err(err).Msg("serf is shutting down")
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
