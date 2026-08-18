package kafka

import (
	"github.com/jonny7/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/internal/core/metrics"
)

type Fake struct {
	metrics metrics.Metrics
}

func (f Fake) Start(done <-chan struct{}, buffer int) chan commander.Event {
	ch := make(chan commander.Event, buffer)
	go func() {
		for {
			select {
			case <-done:
				return
			case ch <- commander.Event{}:
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
