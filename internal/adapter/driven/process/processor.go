package process

import (
	"bytes"
	"net/http"

	"github.com/jonny7/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/internal/core/metrics"
	"github.com/jonny7/maelstrom/internal/logging"
)

type RequestProcessor struct {
	logger  logging.Logger
	metrics metrics.Metrics
}

func NewProcessor(logger logging.Logger, metrics metrics.Metrics) RequestProcessor {
	return RequestProcessor{logger: logger, metrics: metrics}
}

func (r RequestProcessor) Process(done chan struct{}, work <-chan commander.Event, host string, buffer int) chan *http.Request {
	ch := make(chan *http.Request, buffer)
	go func() {
		defer close(ch)
		for {
			select {
			case msg := <-work:

				b := msg.Value

				req, err := http.NewRequest(http.MethodPost, host, bytes.NewBuffer(b))
				if err != nil {
					r.logger.LogWithError(logging.ErrorLevel, "unable to create request", err)
					continue
				}
				if req == nil {
					// handle closed chan
					r.logger.Log(logging.ErrorLevel, "req==nil")
					continue
				}
				select {
				case ch <- req:
					r.metrics.Processed()
				case <-done:
					r.logger.Log(logging.InfoLevel, "closing custom processor")
					return
				}
			case <-done:
				r.logger.Log(logging.InfoLevel, "closing custom processor")
				return
			}
		}
	}()
	return ch
}
