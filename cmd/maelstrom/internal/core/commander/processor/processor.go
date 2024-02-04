package processor

import (
	"net/http"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/requester"
)

//go:generate mockgen -source=processor.go -destination mocks/mock_processor.go -package mocks

// Processor deserializes and converts events to http requests
type Processor interface {
	Process(done chan struct{}, work <-chan requester.Event, host string) chan *http.Request
}
