package discovery

import (
	"github.com/hashicorp/serf/serf"
	"log"
	"net"
)

// Group provides serf event handling hooks
type Group interface {
	Join(name, addr string) error
	Leave(name string) error
}

// Config provides the Node details for each member
type Config struct {
	NodeName       string
	BindAddr       string
	Tags           map[string]string
	StartJoinAddrs []string
}

// Membership provides additional serf.Serf and serf.Event capabilities
type Membership struct {
	Config
	group  Group
	serf   *serf.Serf
	events chan serf.Event
}

// newSerf creates the initial serf.Serf configuration for the member
func (m *Membership) newSerf() error {
	addr, err := net.ResolveTCPAddr("tcp", m.BindAddr)
	if err != nil {
		return err
	}
	// create sane serf defaults to start with
	config := serf.DefaultConfig()
	config.Init()
	// apply member list updates from agent config
	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port
	// create event channel
	config.EventCh = make(chan serf.Event)
	// apply membership tags
	config.Tags = m.Tags
	config.NodeName = m.Config.NodeName
	// finally, create Serf
	m.serf, err = serf.Create(config)
	if err != nil {
		return err
	}
	// start event handler
	go m.eventHandler()
	if m.StartJoinAddrs != nil {
		_, err = m.serf.Join(m.StartJoinAddrs, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func New(group Group, config Config) (*Membership, error) {
	c := &Membership{
		Config: config,
		group:  group,
	}
	if err := c.newSerf(); err != nil {
		return nil, err
	}
	return c, nil
}

func (m *Membership) eventHandler() {
	for e := range m.events {
		// @todo enhance this when eventing is ready
		log.Println(e.EventType().String())
	}
}

func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}
