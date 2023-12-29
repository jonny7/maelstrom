package agent

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

type Agent struct {
	// Config is the agents configuration
	Config Config
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

func (a *Agent) Members() []serf.Member {
	return a.membership.Members()
}

// New returns a new agent or errors. The main configuration is provided through environment vars or defaults.
// The agent will also set up all membership for the Serf cluster
func New(logger zerolog.Logger, consumer commander.Consumer) (*Agent, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	// create agent
	a := &Agent{
		Config:    cfg,
		shutdowns: make(chan struct{}),
	}

	// create commander
	cmdr, err := commander.NewCommander(cfg.Commander, consumer)
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
	return a, nil
}

// setupMembership configures the membership for this Serf node or errors
func (a *Agent) setupMembership(logger zerolog.Logger) error {
	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}

	a.membership, err = membership.New(a.commander, membership.Config{
		NodeName: a.Config.NodeName,
		BindAddr: a.Config.BindAddr,
		Tags: map[string]string{
			"rpc_addr": rpcAddr,
		},
		StartJoinAddrs: a.Config.StartJoinAddrs,
	}, logger)
	return err
}

// serve serves multiplexed TCP and UDP protocols across a single port
func (a *Agent) serve() error {
	if err := a.mux.Serve(); err != nil {
		log.Error().Err(err).Msg("serf is shutting down")
		return err
	}
	return nil
}

// setupMux creates a new connection multiplexer
func (a *Agent) setupMux() error {
	addr, err := net.ResolveTCPAddr("tcp", a.Config.BindAddr)
	if err != nil {
		return err
	}
	rpcAddr := fmt.Sprintf("%s:%d", addr.IP.String(), a.Config.RPCPort)
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}
	a.mux = cmux.New(ln)
	return nil
}
