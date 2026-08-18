package commander

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	mocklogger "github.com/jonny7/maelstrom/cmd/maelstrom/common/logging/mocks"
	mockgenerator "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/generator/mocks"
	mockprocessor "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor/mocks"
	mocksender "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender/mocks"
	mockmetrics "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics/mocks"
)

func newConfig(t *testing.T) Config {
	t.Helper()
	return Config{}
}

func TestNewCommander(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockgenerator.NewMockGenerator(ctrl)
	p := mockprocessor.NewMockProcessor(ctrl)
	h := mocksender.NewMockHTTPDoer(ctrl)
	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

	_ = NewCommander(newConfig(t), c, p, h, l, m)
}

func TestCommanderStop(t *testing.T) {
	done := make(chan struct{})
	u := uuid.New()
	cmdr := Commander{
		runs: map[uuid.UUID]*run{
			u: {done: done},
		},
	}

	cmdr.Stop(u)
	if _, ok := <-done; ok {
		t.Errorf("expected closed channel, but got: %v", done)
	}
	if !cmdr.runs[u].vortex.Finished() {
		t.Error("expected stopped run to be marked finished")
	}
	// a second Stop must be a no-op, not a panic on the closed channel
	cmdr.Stop(u)
}

func TestCommanderStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockgenerator.NewMockGenerator(ctrl)
	p := mockprocessor.NewMockProcessor(ctrl)
	h := mocksender.NewMockHTTPDoer(ctrl)
	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

	c.EXPECT().Start(gomock.Any(), gomock.Any()).Times(1)
	p.EXPECT().Process(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	cmdr := NewCommander(newConfig(t), c, p, h, l, m)
	v := cmdr.Start("http://localhost:8000", 1, 10, 20, 0)

	if v.ID == uuid.Nil {
		t.Error("expected run id to be set")
	}
	if v.Finished() {
		t.Error("expected new run to not be finished")
	}
	if len(cmdr.Vortexes()) != 1 {
		t.Errorf("expected 1 tracked run, got %d", len(cmdr.Vortexes()))
	}
}

func TestAnalytics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockgenerator.NewMockGenerator(ctrl)
	p := mockprocessor.NewMockProcessor(ctrl)
	h := mocksender.NewMockHTTPDoer(ctrl)
	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

	m.EXPECT().Response(gomock.Any()).MinTimes(1)

	done := make(chan struct{})
	results := make(chan result, 10)

	cmdr := NewCommander(newConfig(t), c, p, h, l, m)

	go func() {
		for i := 0; i < 10; i++ {
			results <- result{
				err:      nil,
				response: &http.Response{},
			}
		}
	}()

	cmdr.analytics(done, results)
	time.Sleep(500 * time.Millisecond)
	close(done)
	time.Sleep(500 * time.Millisecond)
}

func TestVortexer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mockgenerator.NewMockGenerator(ctrl)
	p := mockprocessor.NewMockProcessor(ctrl)
	h := mocksender.NewMockHTTPDoer(ctrl)
	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

	h.EXPECT().Do(gomock.Any()).MinTimes(1)
	m.EXPECT().Requested().MinTimes(1)
	l.EXPECT().Log(gomock.Any(), gomock.Any()).MinTimes(1)

	done := make(chan struct{})
	work := make(chan *http.Request)

	cmdr := NewCommander(newConfig(t), c, p, h, l, m)

	go func() {
		for i := 0; i < 10; i++ {
			r, _ := http.NewRequest(http.MethodGet, "localhost:8080", nil)
			work <- r
		}
	}()

	cmdr.vortexer(done, work, 5)
	time.Sleep(500 * time.Millisecond)
	close(done)
	time.Sleep(500 * time.Millisecond)
}
