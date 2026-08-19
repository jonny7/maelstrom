package prometheus

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/jonny7/maelstrom/internal/core/metrics"
)

type Engine struct {
	consumed  prometheus.Counter
	processed prometheus.Counter
	requested prometheus.Counter
	response  *prometheus.CounterVec
}

func NewMetrics() metrics.Metrics {
	e := Engine{
		consumed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "maelstrom_consumed_events",
			Help: "Number of events consumed from Kafka",
		}),
		processed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "maelstrom_processed_events",
			Help: "Number of processed Kafka events",
		}),
		requested: promauto.NewCounter(prometheus.CounterOpts{
			Name: "maelstrom_requested",
			Help: "Number of requested created to send to target",
		}),
		response: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "maelstrom_responses",
			Help: "Number of responses back from the target by status code",
		}, []string{"status"}),
	}
	return e
}

func (e Engine) Processed() {
	e.processed.Inc()
}

func (e Engine) Requested() {
	e.requested.Inc()
}

func (e Engine) Response(status int) {
	e.response.WithLabelValues(strconv.Itoa(status)).Inc()
}

func (e Engine) Consumed() {
	e.consumed.Inc()
}
