package process

import (
	"bytes"
	"net/http"

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/requester"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

type RequestProcessor struct {
	logger  logging.Logger
	metrics metrics.Metrics
}

func NewProcessor(logger logging.Logger, metrics metrics.Metrics) RequestProcessor {
	return RequestProcessor{logger: logger, metrics: metrics}
}

func (r RequestProcessor) Process(done chan struct{}, work <-chan requester.Event, host string) chan *http.Request {
	ch := make(chan *http.Request)
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
				ch <- req
				r.metrics.Processed()
			case <-done:
				r.logger.Log(logging.InfoLevel, "closing custom processor")
				return
			}
		}
	}()
	return ch
}
