package processor

import (
	"bytes"
	"net/http"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/proto/analytics"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type RequestProcessor struct{}

func (r RequestProcessor) Process(done chan struct{}, work <-chan commander.Event) chan *http.Request {
	ch := make(chan *http.Request)
	go func() {
		defer close(ch)
		for {
			select {
			case msg := <-work:
				var fw analytics.FWRequestEvent
				err := proto.Unmarshal(msg.Value, &fw)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}
				b, err := protojson.Marshal(&fw)
				if err != nil {
					log.Error().Err(err).Send()
					continue
				}

				req, err := http.NewRequest(http.MethodPost, "http://localhost:8000/openrtb2/markets", bytes.NewBuffer(b))
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
