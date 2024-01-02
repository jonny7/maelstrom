package agent

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
	"github.com/travisjeffery/go-dynaport"
)

type dummyConsumer struct{}

func (d dummyConsumer) Consume() {}
func (d dummyConsumer) Close()   {}

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

		if _, err := New(zerolog.Logger{}, dummyConsumer{}); err != nil {
			t.Errorf("expected no error: %v", err)
		}
	}
}
