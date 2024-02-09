package consumer

import "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/requester"

//go:generate mockgen -source=consumer.go -destination mocks/mock_consumer.go -package mocks

// Consumer provides a mechanism to consume events from any source system
type Consumer interface {
	Start(done <-chan struct{}, buffer int) chan requester.Event
	Close()
}
