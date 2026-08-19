// Package serf implements the agent's Membership port over a hashicorp/serf cluster.
package serf

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/serf/serf"

	"github.com/jonny7/maelstrom/internal/core/agent"
)

// Config provides the node details for this member. Parsed from the environment by main.
type Config struct {
	// Hostname is the provided host for this agent, if running in K8s it's equivalent to os.GetEnv("HOSTNAME")
	Hostname string `env:"HOSTNAME" envDefault:"0.0.0.0"`
	SerfPort string `env:"SERF_PORT" envDefault:"7946"`
	// BindAddr is the address serf runs on https://www.serf.io/docs/agent/options.html#ports-used
	BindAddr string `env:"SERF_SERVICE,expand" envDefault:"$HOSTNAME:$SERF_PORT"`
	// NodeName is this node's unique name within the cluster
	NodeName string `env:"NODE_NAME,expand" envDefault:"$HOSTNAME"`
	// RPCPort is advertised to peers in the rpc_addr member tag
	RPCPort int `env:"RPC_PORT" envDefault:"7373"`
	// StartJoinAddrs is a list of seeds, this should be the 1st instance running
	StartJoinAddrs []string `env:"SEED_NODES" envDefault:"maelstrom-0.maelstrom-svc.default.svc.cluster.local:7946"`
}

// Membership implements the agent.Membership port over serf.
type Membership struct {
	serf *serf.Serf
}

// New creates the serf node and, when seeds are configured, joins the cluster.
func New(cfg Config) (*Membership, error) {
	addr, err := net.ResolveTCPAddr("tcp", cfg.BindAddr)
	if err != nil {
		return nil, err
	}
	config := serf.DefaultConfig()
	// keep left/failed nodes visible briefly, then reap
	config.ReconnectTimeout = 5 * time.Minute
	config.TombstoneTimeout = 10 * time.Minute
	config.Init()
	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port
	config.NodeName = cfg.NodeName
	// advertise this node's rpc address to the cluster; the UI shows it on each node card
	host, _, err := net.SplitHostPort(cfg.BindAddr)
	if err != nil {
		return nil, err
	}
	config.Tags = map[string]string{"rpc_addr": fmt.Sprintf("%s:%d", host, cfg.RPCPort)}

	s, err := serf.Create(config)
	if err != nil {
		return nil, err
	}
	// @todo handle bootstrapping
	if cfg.StartJoinAddrs != nil {
		if _, err = s.Join(cfg.StartJoinAddrs, true); err != nil {
			return nil, err
		}
	}
	return &Membership{serf: s}, nil
}

// Members returns the current member list.
func (m *Membership) Members() []agent.Member {
	serfMembers := m.serf.Members()
	members := make([]agent.Member, 0, len(serfMembers))
	for _, sm := range serfMembers {
		members = append(members, agent.Member{
			Name:   sm.Name,
			Addr:   sm.Addr.String(),
			Port:   int(sm.Port),
			Tags:   sm.Tags,
			Status: sm.Status.String(),
		})
	}
	return members
}

// Leave announces this node's departure so peers mark it left instead of failed.
func (m *Membership) Leave() error {
	return m.serf.Leave()
}
