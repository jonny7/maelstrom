package agent

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/travisjeffery/go-dynaport"
	"go.uber.org/mock/gomock"

	mocklogger "github.com/jonny7/maelstrom/cmd/maelstrom/common/logging/mocks"
	mockgenerator "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/generator/mocks"
	mock "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/mocks"
	mockprocessor "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor/mocks"
	mocksender "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender/mocks"
	mockmetrics "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics/mocks"
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
	l.EXPECT().Log(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	l.EXPECT().LogWithError(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	var seed string
	for i := range 3 {
		ports := dynaport.Get(2)
		bindAddr := fmt.Sprintf("127.0.0.1:%d", ports[0])
		if i == 0 {
			seed = bindAddr
		}

		cfg := Config{
			BindAddr:       bindAddr,
			RPCPort:        ports[1],
			NodeName:       fmt.Sprintf("node-%d", i),
			StartJoinAddrs: []string{seed},
		}

		if _, err := New(cfg, l, c, p, h, m); err != nil {
			t.Errorf("expected no error: %v", err)
		}
	}
}
func TestAgentStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmdr := mock.NewMockCommand(ctrl)
	l := mocklogger.NewMockLogger(ctrl)
	l.EXPECT().Log(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	a := agent{
		commander: cmdr,
		logger:    l,
	}
	cmdr.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	a.Start("", 1, 1, 0, 0)
}

func TestAgentStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmdr := mock.NewMockCommand(ctrl)
	l := mocklogger.NewMockLogger(ctrl)
	l.EXPECT().Log(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	a := agent{
		commander: cmdr,
		logger:    l,
	}
	u := uuid.New()
	cmdr.EXPECT().Stop(u)
	a.Stop(u)
}
