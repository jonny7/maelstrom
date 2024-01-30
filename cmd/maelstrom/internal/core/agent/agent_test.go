package agent

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging/mocks"
	mockscommander "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/mocks"
	mocksmetrics "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics/mocks"
	"github.com/travisjeffery/go-dynaport"
	"go.uber.org/mock/gomock"
)

func TestAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := mockscommander.NewMockConsumer(ctrl)
	processor := mockscommander.NewMockProcessor(ctrl)
	h := mockscommander.NewMockHTTPClient(ctrl)
	m := mocksmetrics.NewMockMetrics(ctrl)
	l := mocks.NewMockLogger(ctrl)

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

		if err := os.Setenv("NODE_NAME", fmt.Sprintf("0.0.0.0%s", os.Getenv("SERF_SERVICE,expand"))); err != nil {
			t.Error(err)
		}

		if i == 0 {
			if err := os.Setenv("SEED_NODES", fmt.Sprintf("0.0.0.0:%d", ports[0])); err != nil {
				t.Error(err)
			}
		}

		if _, err := New(l, consumer, processor, h, m); err != nil {
			t.Errorf("expected no error: %v", err)
		}
	}
}
