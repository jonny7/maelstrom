package agent

import (
	"fmt"
	"net"
)

type Config struct {
	// Hostname is the provided host for this agent, if running in K8s it's equivalent to os.GetEnv("HOSTNAME")
	Hostname string `env:"HOSTNAME" envDefault:"0.0.0.0"`
	SerfPort string `env:"SERF_PORT" envDefault:"7946"`
	// BindAddr is the address serf runs on https://www.serf.io/docs/agent/options.html#ports-used
	BindAddr string `env:"SERF_SERVICE,expand" envDefault:"$HOSTNAME:$SERF_PORT"`
	// RPCPort is the port for client connections https://www.serf.io/docs/agent/options.html#ports-used
	RPCPort int `env:"RPC_PORT" envDefault:"7373"`
	// NodeName is this node's unique name within the cluster
	NodeName string `env:"NODE_NAME,expand" envDefault:"$HOSTNAME"`
	// StartJoinAddrs is a list of seeds, this should be the 1st instance running
	StartJoinAddrs []string `env:"SEED_NODES" envDefault:"maelstrom-0.maelstrom-svc.default.svc.cluster.local:7946"`
}

// RPCAddr splits the provided BindAddress to the HOST and combines the RPC_PORT
func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}
