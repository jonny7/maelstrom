package commander

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	mockmetrics "github.com/jonny7/maelstrom/internal/core/metrics/mocks"
	mocklogger "github.com/jonny7/maelstrom/internal/logging/mocks"
)

// doerFunc stubs HTTPDoer; the generated mock lives in commander/mocks which an
// in-package test cannot import without a cycle.
type doerFunc func(req *http.Request) (*http.Response, error)

func (d doerFunc) Do(req *http.Request) (*http.Response, error) { return d(req) }

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

func TestAnalytics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mockmetrics.NewMockMetrics(ctrl)
	m.EXPECT().Response(gomock.Any()).MinTimes(1)

	done := make(chan struct{})
	results := make(chan result, 10)

	cmdr := Commander{metrics: m}

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

	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

	m.EXPECT().Requested().MinTimes(1)
	l.EXPECT().Log(gomock.Any(), gomock.Any()).MinTimes(1)

	var calls atomic.Int32
	h := doerFunc(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200}, nil
	})

	done := make(chan struct{})
	work := make(chan *http.Request)

	cmdr := Commander{client: h, logger: l, metrics: m}

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

	if calls.Load() == 0 {
		t.Error("expected the client to have been called")
	}
}
