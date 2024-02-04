package metrics

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

type Service struct {
	metrics.Metrics
}

func NewMetrics(m metrics.Metrics) Service {
	return Service{m}
}
