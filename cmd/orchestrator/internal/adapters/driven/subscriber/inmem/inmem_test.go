package inmem

import (
	"testing"
)

func TestInMemorySubscriberPush(t *testing.T) {
	inmem := NewInMemoryRepository()
	if err := inmem.Add("127.0.0.1"); err != nil {
		t.Error(err)
	}
	nodeListLength := len(inmem.nodeList)
	if nodeListLength != 1 {
		t.Errorf("want 1, got %d", nodeListLength)
	}
}
