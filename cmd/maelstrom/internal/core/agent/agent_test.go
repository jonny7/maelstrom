package agent

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/rs/zerolog"
	"github.com/travisjeffery/go-dynaport"
)

var serverURL string

type dummyConsumer struct{}

func (d dummyConsumer) Consume() chan commander.Event {
	ch := make(chan commander.Event)
	go func() {
		defer close(ch)
		ch <- commander.Event{}
	}()
	return ch
}
func (d dummyConsumer) Close() {}

type processor struct{}

func (p processor) Process(_ <-chan commander.Event) chan *http.Request {
	ch := make(chan *http.Request)
	go func() {
		defer close(ch)
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080", nil)
		ch <- req
	}()
	return ch
}

func TestAgent(t *testing.T) {
	for i := 0; i < 3; i++ {
		ports := dynaport.Get(2)

		if err := os.Setenv("SERF_PORT", strconv.Itoa(ports[0])); err != nil {
			t.Error(err)
		}

		if err := os.Setenv("SERF_SERVICE,expand", fmt.Sprintf(":%d", ports[0])); err != nil {
			t.Error(err)
		}

		if err := os.Setenv("RPC_PORT", strconv.Itoa(ports[1])); err != nil {
			t.Error(err)
		}

		if i == 0 {
			if err := os.Setenv("SEED_NODES", fmt.Sprintf("0.0.0.0:%d", ports[0])); err != nil {
				t.Error(err)
			}
		}

		if _, err := New(zerolog.Logger{}, dummyConsumer{}, processor{}); err != nil {
			t.Errorf("expected no error: %v", err)
		}
	}
}
