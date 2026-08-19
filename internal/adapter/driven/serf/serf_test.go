package serf_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"

	"github.com/jonny7/maelstrom/internal/adapter/driven/serf"
)

func TestMembership(t *testing.T) {
	var members []*serf.Membership
	var seed string
	rpcAddrs := map[string]string{}

	for i := range 3 {
		ports := dynaport.Get(2)
		addr := fmt.Sprintf("127.0.0.1:%d", ports[0])
		cfg := serf.Config{
			NodeName: fmt.Sprintf("%d", i),
			BindAddr: addr,
			RPCPort:  ports[1],
		}
		rpcAddrs[cfg.NodeName] = fmt.Sprintf("127.0.0.1:%d", ports[1])
		if i == 0 {
			seed = addr
		} else {
			cfg.StartJoinAddrs = []string{seed}
		}

		m, err := serf.New(cfg)
		require.NoError(t, err)
		members = append(members, m)
	}

	require.Eventually(t, func() bool {
		return len(members[0].Members()) == 3
	}, 3*time.Second, 250*time.Millisecond)

	// every member's rpc_addr tag is gossiped across the cluster
	for _, m := range members[0].Members() {
		require.Equal(t, rpcAddrs[m.Name], m.Tags["rpc_addr"])
	}

	require.NoError(t, members[2].Leave())

	require.Eventually(t, func() bool {
		for _, m := range members[0].Members() {
			if m.Name == "2" {
				return m.Status == "left"
			}
		}
		return false
	}, 7*time.Second, 250*time.Millisecond)
}
