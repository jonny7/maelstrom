package membership_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/rs/zerolog"
	"github.com/travisjeffery/go-dynaport"

	. "github.com/jonny7/maelstrom/cmd/maelstrom/internal/membership"
	"github.com/stretchr/testify/require"
)

// member implements membership interface
type member struct {
	joins  chan map[string]string
	leaves chan string
}

func (m *member) Join(id, addr string) error {
	if m.joins != nil {
		m.joins <- map[string]string{
			"id":   id,
			"addr": addr,
		}
	}
	return nil
}

func (m *member) Leave(id string) error {
	if m.leaves != nil {
		m.leaves <- id
	}
	return nil
}

func TestMembership(t *testing.T) {
	members, m := setupMembership(t, nil)
	members, _ = setupMembership(t, members)
	members, _ = setupMembership(t, members)

	require.Eventually(t, func() bool {
		return 2 == len(m.joins) &&
			3 == len(members[0].Members()) &&
			0 == len(m.leaves)
	}, 3*time.Second, 250*time.Millisecond)

	require.NoError(t, members[2].Leave())

	require.Eventually(t, func() bool {
		return 2 == len(m.joins) &&
			3 == len(members[0].Members()) &&
			serf.StatusLeft == members[0].Members()[2].Status &&
			1 == len(m.leaves)
	}, 7*time.Second, 250*time.Millisecond)

	require.Equal(t, fmt.Sprintf("%d", 2), <-m.leaves)
}

func setupMembership(t *testing.T, members []*Membership) ([]*Membership, *member) {
	id := len(members)
	ports := dynaport.Get(1)
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
	tags := map[string]string{
		"rpc_addr": addr,
	}
	c := Config{
		NodeName: fmt.Sprintf("%d", id),
		BindAddr: addr,
		Tags:     tags,
	}
	newMember := &member{}
	if len(members) == 0 {
		newMember.joins = make(chan map[string]string, 3)
		newMember.leaves = make(chan string, 3)
	} else {
		c.StartJoinAddrs = []string{
			members[0].BindAddr,
		}
	}
	node, err := New(newMember, c, zerolog.Logger{})
	require.NoError(t, err)
	members = append(members, node)
	return members, newMember
}
