package commander

import (
	"net/http"
	"time"
)

//go:generate mockgen -source=ports.go -destination mocks/ports.go -package mock

// Event is a single unit of load, produced by a Generator and turned into an HTTP request by a Processor.
type Event struct {
	Key       []byte
	Value     []byte
	Headers   []Header
	Timestamp time.Time
}

type Header struct {
	Key   string
	Value []byte
}

// Generator provides a mechanism to generate data to load test. It is the initial starting point for Maelstrom.
// An example could be a Kafka stream, or a simple function.
type Generator interface {
	Start(done <-chan struct{}, buffer int) chan Event
	Close()
}

// Processor deserializes and converts events to http requests
type Processor interface {
	Process(done chan struct{}, work <-chan Event, host string, buffer int) chan *http.Request
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
