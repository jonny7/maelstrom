package agent

import (
	"fmt"
	"net"

	"github.com/caarlos0/env/v10"
	"github.com/hashicorp/serf/serf"
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/consumer"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/membership"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
	"github.com/rs/zerolog/log"
	"github.com/soheilhy/cmux"
)

type agent struct {
	// Config is the agents configuration
	config Config
	// Commander is a raft based set of changes to apply across the Maelstrom nodes
	commander commander.Commander
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
	Start(host string, jobs, workers int, cbuf, rbuf int)
	Stop()
}

func (a Agent) Start(host string, jobs, workers int, cbuf, rbuf int) {
	log.Debug().Msg("starting consumer and processor")
	a.agent.commander.Start(host, jobs, workers, cbuf, rbuf)
}

func (a Agent) Stop() {
	log.Debug().Msg("stop load test was triggered by the user")
	a.agent.commander.Stop()
}

func (a Agent) Members() []serf.Member {
	return a.agent.membership.Members()
}

// New returns a new agent or errors. The main configuration is provided through environment vars or defaults.
// The agent will also set up all membership for the Serf cluster
func New(logger logging.Logger, consumer consumer.Consumer, processor processor.Processor, client sender.HTTPDoer, metrics metrics.Metrics) (Service, error) {
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
	a.commander = commander.NewCommander(cfg.Commander, consumer, processor, client, logger, metrics)

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
	return &Agent{agent: a}, nil
}

// setupMembership configures the membership for this Serf node or errors
func (a *agent) setupMembership(logger logging.Logger) error {
	rpcAddr, err := a.config.RPCAddr()
	if err != nil {
		return err
	}

	a.membership, err = membership.New(&a.commander, membership.Config{
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
