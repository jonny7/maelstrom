package inmem

import (
	"sync"
)

type InMemorySubscriber struct {
	nodeList []string
	mu       sync.Mutex
}

func NewInMemoryRepository() *InMemorySubscriber {
	return &InMemorySubscriber{}
}

func (i *InMemorySubscriber) Pop(addr string) error {
	return nil
}

func (i *InMemorySubscriber) Push(addr string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.nodeList = append(i.nodeList, addr)
	return nil
}

func (i *InMemorySubscriber) List() []string {
	return i.nodeList
}
