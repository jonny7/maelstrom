package kafka

import (
	"log"
	"time"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/requester"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

type Fake struct {
	metrics metrics.Metrics
}

func (f Fake) Start(done <-chan struct{}) chan requester.Event {
	ch := make(chan requester.Event)
	go func() {
		var i int
		for {
			select {
			case <-done:
				return
			default:
				time.Sleep(time.Millisecond * 20)
				i++
				log.Println(i, ":", time.Now().Unix())
				ch <- requester.Event{}
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
