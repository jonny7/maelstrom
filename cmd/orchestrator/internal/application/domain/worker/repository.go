package worker

import (
	"fmt"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/adapters/driven/subscriber/inmem"
)

type Store interface {
	Add(worker string) error
	Remove(worker string) error
	List() []string
}

func newSubscriberStore(method SubscriberStore) (Store, error) {
	switch method {
	case InMemory:
		return inmem.NewInMemoryRepository(), nil
	default:
		return nil, fmt.Errorf("unrecognized subscriber method")
	}
}
