package metrics

import "github.com/jonny7/maelstrom/internal/core/metrics"

type Nop struct{}

func (n Nop) Consumed() {}

func (n Nop) Processed() {}

func (n Nop) Requested() {}

func (n Nop) Response(_ int) {}

func (n Nop) Increment() {}

func NewNop() metrics.Metrics {
	return Nop{}
}
