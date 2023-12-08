package commander

import (
	"fmt"

	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
)

type Commander struct {
	config Config
	raft   *raft.Raft
}

// Leave returns the attempted raft removal of the node
func (c *Commander) Leave(id string) error {
	return c.raft.RemoveServer(raft.ServerID(id), 0, 0).Error()
}

// Join returns the attempt to add a new voter to the cluster
func (c *Commander) Join(id, addr string) error {
	return nil
}

func NewCommander(cfg Config) (*Commander, error) {
	cmdr := &Commander{
		config: cfg,
	}
	return cmdr, nil
}

func newRaft() {
	cfg := raft.DefaultConfig()
	fmt.Print(cfg)
	_, err := boltdb.New(boltdb.Options{})
	if err != nil {
		return
	}
}
