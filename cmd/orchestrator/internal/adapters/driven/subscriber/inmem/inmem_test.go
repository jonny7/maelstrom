package inmem

import (
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/domain/worker"
	"testing"
)

const ip = "127.0.0.1"

func TestInMemoryAdd(t *testing.T) {
	inmem := NewInMemoryRepository()
	if err := inmem.Add(worker.Worker{IP: ip}); err != nil {
		t.Error(err)
	}

	if len(inmem.nodeList) != 1 {
		t.Errorf("want 1, got %d", len(inmem.nodeList))
	}
}

func TestInMemoryRemove(t *testing.T) {
	inmem := NewInMemoryRepository()
	if err := inmem.Add(worker.Worker{IP: ip}); err != nil {
		t.Error(err)
	}
	if err := inmem.Remove(worker.Worker{IP: ip}); err != nil {
		t.Error(err)
	}
	if len(inmem.nodeList) != 0 {
		t.Errorf("want 0, but got %d", len(inmem.nodeList))
	}
}
