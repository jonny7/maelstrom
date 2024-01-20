package processor

import (
	"bytes"
	"net/http"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/rs/zerolog/log"
)

type RequestProcessor struct{}

func (r RequestProcessor) Process(done chan struct{}, work <-chan commander.Event) chan *http.Request {
	ch := make(chan *http.Request)
	go func() {
		defer close(ch)
		for {
			select {
			case msg := <-work:
				// @todo proper protobuf decode
				decode := msg
				req, err := http.NewRequest(http.MethodPost, "http://localhost:55000/", bytes.NewBuffer(decode.Value))
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
