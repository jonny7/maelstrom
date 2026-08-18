package kafka

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/requester"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

type Fake struct {
	metrics metrics.Metrics
}

func (f Fake) Start(done <-chan struct{}, buffer int) chan requester.Event {
	ch := make(chan requester.Event, buffer)
	go func() {
		for {
			select {
			case <-done:
				return
			case ch <- requester.Event{}:
				f.metrics.Consumed()
			}
		}
	}()
	return ch
}

func (f Fake) Close() {}

func NewFake(metrics metrics.Metrics) (Fake, error) {
	return Fake{metrics: metrics}, nil
}
