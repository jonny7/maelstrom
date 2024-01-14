package processor

import (
	"bytes"
	"net/http"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/rs/zerolog/log"
)

type RequestProcessor struct{}

func (r RequestProcessor) Process(work <-chan commander.Event) chan *http.Request {
	ch := make(chan *http.Request)
	go func() {
		defer close(ch)
		for {
			select {
			case msg := <-work:
				// @todo proper protobuf decode
				decode := msg
				req, err := http.NewRequest(http.MethodPost, "https://localhost:8080", bytes.NewBuffer(decode.Value))
				if err != nil {
					log.Error().Err(err).Send()
				}
				ch <- req
			}
		}
	}()
	return ch
}
