package generator

import "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/requester"

//go:generate mockgen -source=generator.go -destination mocks/mock_generator.go -package mocks

// Generator provides a mechanism to generate data to load test. It is the initial starting point for Maelstrom
// an example could a Kafka stream, or a simple function.
type Generator interface {
	Start(done <-chan struct{}, buffer int) chan requester.Event
	Close()
}
