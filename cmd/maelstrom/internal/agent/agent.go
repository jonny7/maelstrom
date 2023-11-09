package agent

import (
	"fmt"
	"github.com/caarlos0/env/v9"
	"github.com/hashicorp/serf/serf"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/discovery"
	"github.com/soheilhy/cmux"
	"log"
	"net"
	"sync"
)

type Config struct {
	// Hostname is the provided host for this agent, if running in K8s it's equivalent to os.GetEnv("HOSTNAME")
	Hostname string `env:"HOSTNAME" envDefault:"0.0.0.0"`
	// DataDir stores raft data.
	DataDir string `env:"DATA_DIR" envDefault:"data"`
	// BindAddr is the address serf runs on https://www.serf.io/docs/agent/options.html#ports-used
	BindAddr string `env:"SERF_SERVICE,expand" envDefault:"$HOSTNAME:7946"`
	// RPCPort is the port for client (and Raft) connections https://www.serf.io/docs/agent/options.html#ports-used
	RPCPort int `env:"RPC_PORT" envDefault:"7373"`
	// Raft server id.
	NodeName string `env:"NODE_NAME,expand" envDefault:"$HOSTNAME"`
	// StartJoinAddrs is a list of seeds, this should be the 1st instance running
	StartJoinAddrs []string `env:"SEED_NODES" envDefault:"maelstrom-0.maelstrom-svc.default.svc.cluster.local:7946"`
	// Bootstrap should be set to true when starting the first node of the cluster. You probably only need this in non-k8s environments
	Bootstrap bool `env:"BOOTSTRAP" envDefault:"false"`
}

type Agent struct {
	// Config is the agents configuration
	Config Config
	// Mux helps serve UDP & TCP over the same port
	Mux cmux.CMux
	// membership configures Serf and eventing
	membership *discovery.Membership
	// shutdowns receive channel events, signifying it should be shut down
	shutdowns    chan struct{}
	shutdownLock sync.Mutex
}

func (a *Agent) Members() []serf.Member {
	return a.membership.Members()
}

// RPCAddr splits the provided BindAddress to the HOST and combines the RPC_PORT
func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}

// New returns a new agent or errors. The main configuration is provided through environment vars or defaults.
// The agent will also set up all membership for the Serf cluster
func New() (*Agent, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	// create agent
	a := &Agent{
		Config:    cfg,
		shutdowns: make(chan struct{}),
	}

	// setup membership or err
	if err := a.setupMembership(); err != nil {
		return nil, err
	}
	// setup mux or err
	if err := a.setupMux(); err != nil {
		return nil, err
	}
	// serve mux @todo handle err
	go a.serve()
	return a, nil
}

// @todo replace this will application structure
type b struct{}

func (b b) Join(name, addr string) error {
	log.Println(name, addr, "joining")
	return nil
}

func (b b) Leave(name string) error {
	log.Println(name, "joining")
	return nil
}

// setupMembership configures the membership for this Serf node or errors
func (a *Agent) setupMembership() error {
	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}
	a.membership, err = discovery.New(b{}, discovery.Config{
		NodeName: a.Config.NodeName,
		BindAddr: a.Config.BindAddr,
		Tags: map[string]string{
			"rpc_addr": rpcAddr,
		},
		StartJoinAddrs: a.Config.StartJoinAddrs,
	})
	return err
}

// serve serves multiplexed TCP and UDP protocols across a single port
func (a *Agent) serve() error {
	if err := a.Mux.Serve(); err != nil {
		log.Println("shutting down serf ports")
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
	a.Mux = cmux.New(ln)
	return nil
}
