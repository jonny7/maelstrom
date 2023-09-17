package worker

import (
	"fmt"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/adapters/driven/subscriber/inmem"
)

type Subscriber interface {
	// Subscribe()
	Pop(remoteAddr string) error
	Push(remoteAddr string) error
	List() []string
}

func NewSubscriber(method SubscriberStore) (Subscriber, error) {
	switch method {
	case InMemory:
		return inmem.NewInMemoryRepository(), nil
	default:
		return nil, fmt.Errorf("unrecognized subscriber method")
	}
}
