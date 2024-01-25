package processor

import (
	"bytes"
	"net/http"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/rs/zerolog/log"
)

type RequestProcessor struct{}

func NewProcessor() RequestProcessor {
	return RequestProcessor{}
}

func (r RequestProcessor) Process(done chan struct{}, work <-chan commander.Event, host string) chan *http.Request {
	ch := make(chan *http.Request)
	go func() {
		defer close(ch)
		for {
			select {
			case msg := <-work:
				b := msg.Value
				// @todo env var for host or override in this func https://conduit-server-internal.use1.dev.aws.viacbs.tech
				req, err := http.NewRequest(http.MethodPost, host, bytes.NewBuffer(b))
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}
				ch <- req
			case <-done:
				log.Debug().Msg("closing custom processing")
				return
			}
		}
	}()
	return ch
}
