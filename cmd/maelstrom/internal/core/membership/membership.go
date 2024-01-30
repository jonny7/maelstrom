package membership

import (
	"log"
	"net"

	"github.com/hashicorp/serf/serf"
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
)

// Member must provide leaving and joining events
type Member interface {
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
	member Member
	serf   *serf.Serf
	events chan serf.Event
	logger logging.Logger
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
	m.events = make(chan serf.Event)
	config.EventCh = m.events
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
	// if we're not the first node to join then attempt to join the seed nodes
	// @todo handle bootstrapping
	if m.StartJoinAddrs != nil {
		_, err = m.serf.Join(m.StartJoinAddrs, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// New creates a new Member with the provided Group and Config
func New(member Member, config Config, baseLogger logging.Logger) (*Membership, error) {
	c := &Membership{
		Config: config,
		member: member,
		logger: baseLogger,
	}
	if err := c.newSerf(); err != nil {
		return nil, err
	}
	return c, nil
}

// eventHandler processes all serf events for a node.
func (m *Membership) eventHandler() {
	for e := range m.events {
		log.Println(e.EventType().String())
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					return
				}
				m.handleLeave(member)
			}
		default:
			// @todo
			m.logger.Log(logging.InfoLevel, "handling default event", logging.KV{
				Key:   "event type",
				Value: e.String(),
			})
		}
	}
}

// Members returns the current member list
func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}

// Leave allows a member to gracefully leave the member list
func (m *Membership) Leave() error {
	return m.serf.Leave()
}

// isLocal determines if this event (most likely) is occurring on the local node
func (m *Membership) isLocal(member serf.Member) bool {
	return m.serf.LocalMember().Name == member.Name
}

// handleJoin attempts to add a member to an existing cluster
func (m *Membership) handleJoin(member serf.Member) {
	if err := m.member.Join(
		member.Name,
		member.Tags["rpc_addr"],
	); err != nil {
		m.logger.LogWithError(logging.ErrorLevel, "failed to join cluster", err)
	}
}

// handleLeave attempts to remove a member from an existing cluster
func (m *Membership) handleLeave(member serf.Member) {
	if err := m.member.Leave(
		member.Name,
	); err != nil {
		m.logger.LogWithError(logging.ErrorLevel, "failed to leave cluster", err)
	}
}
