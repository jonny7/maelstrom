package inmem

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/domain/worker"
	"slices"
	"sync"
)

type InMemoryWorkerStore struct {
	nodeList []worker.Worker
	mu       sync.Mutex
}

func NewInMemoryRepository() *InMemoryWorkerStore {
	return &InMemoryWorkerStore{}
}

func (i *InMemoryWorkerStore) Remove(addr worker.Worker) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.nodeList = slices.DeleteFunc(i.nodeList, func(w worker.Worker) bool {
		return w == addr
	})
	return nil
}

func (i *InMemoryWorkerStore) Add(addr worker.Worker) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.nodeList = append(i.nodeList, addr)
	return nil
}

func (i *InMemoryWorkerStore) List() []worker.Worker {
	return i.nodeList
}
