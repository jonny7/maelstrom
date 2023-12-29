package commander

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
)

type Commander struct {
	config   Config
	raft     *raft.Raft
	consumer Consumer
	client   http.Client
}

type Consumer interface {
	Consume()
	Close()
}

// Leave returns the attempted raft removal of the node
func (c *Commander) Leave(id string) error {
	return c.raft.RemoveServer(raft.ServerID(id), 0, 0).Error()
}

// Join returns the attempt to add a new voter to the cluster
func (c *Commander) Join(id, addr string) error {
	return nil
}

func NewCommander(cfg Config, consumer Consumer) (*Commander, error) {
	cmdr := &Commander{
		config:   cfg,
		consumer: consumer,
	}

	go func() {
		consumer.Consume()
	}()

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
