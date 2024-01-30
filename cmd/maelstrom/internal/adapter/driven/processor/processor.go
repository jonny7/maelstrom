package processor

import (
	"bytes"
	"net/http"

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
)

type RequestProcessor struct {
	logger logging.Logger
}

func NewProcessor(logger logging.Logger) RequestProcessor {
	return RequestProcessor{logger: logger}
}

func (r RequestProcessor) Process(done chan struct{}, work <-chan commander.Event, host string) chan *http.Request {
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
					continue
				}
				ch <- req
			case <-done:
				r.logger.Log(logging.InfoLevel, "closing custom processor")
				return
			}
		}
	}()
	return ch
}
