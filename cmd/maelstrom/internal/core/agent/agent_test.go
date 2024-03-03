package agent

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/google/uuid"
	mocklogger "github.com/jonny7/maelstrom/cmd/maelstrom/common/logging/mocks"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	mockgenerator "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/generator/mocks"
	mock "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/mocks"
	mockprocessor "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor/mocks"
	mocksender "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender/mocks"
	mockmetrics "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics/mocks"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/travisjeffery/go-dynaport"
	"go.uber.org/mock/gomock"
)

func TestRPCAddrErr(t *testing.T) {
	c := Config{BindAddr: "localhost"}
	_, err := c.RPCAddr()
	if err == nil {
		t.Errorf("invalid host:port. got %s", c.BindAddr)
	}
}

func TestAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockgenerator.NewMockGenerator(ctrl)
	p := mockprocessor.NewMockProcessor(ctrl)
	h := mocksender.NewMockHTTPDoer(ctrl)
	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

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

		if _, err := New(l, c, p, h, m); err != nil {
			t.Errorf("expected no error: %v", err)
		}
	}
}
func TestAgentStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmdr := mock.NewMockCommand(ctrl)
	a := agent{
		commander: cmdr,
	}
	cmdr.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	a.Start("", 1, 1, 0, 0)
}

func TestAgentStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmdr := mock.NewMockCommand(ctrl)
	runs := make([]maelstrom.Vortex, 0)
	runs = append(runs, maelstrom.Vortex{Id: commander.ToString(uuid.NewString())})
	a := agent{
		commander: cmdr,
		runs:      runs,
	}
	cmdr.EXPECT().Stop(gomock.Any())
	a.Stop(uuid.MustParse(*runs[0].Id))
}
