package commander_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/jonny7/maelstrom/internal/core/commander"
	mock "github.com/jonny7/maelstrom/internal/core/commander/mocks"
	mockmetrics "github.com/jonny7/maelstrom/internal/core/metrics/mocks"
	mocklogger "github.com/jonny7/maelstrom/internal/logging/mocks"
)

func TestNewCommander(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock.NewMockGenerator(ctrl)
	p := mock.NewMockProcessor(ctrl)
	h := mock.NewMockHTTPDoer(ctrl)
	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

	_ = commander.NewCommander(c, p, h, l, m)
}

func TestCommanderStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock.NewMockGenerator(ctrl)
	p := mock.NewMockProcessor(ctrl)
	h := mock.NewMockHTTPDoer(ctrl)
	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

	c.EXPECT().Start(gomock.Any(), gomock.Any()).Times(1)
	p.EXPECT().Process(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	cmdr := commander.NewCommander(c, p, h, l, m)
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

// TestCommanderStartStop drives a full run lifecycle with traffic; -race and the
// guarded channel sends are what this is exercising.
func TestCommanderStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c := mock.NewMockGenerator(ctrl)
	p := mock.NewMockProcessor(ctrl)
	h := mock.NewMockHTTPDoer(ctrl)
	m := mockmetrics.NewMockMetrics(ctrl)
	l := mocklogger.NewMockLogger(ctrl)

	l.EXPECT().Log(gomock.Any(), gomock.Any()).AnyTimes()
	l.EXPECT().Log(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	m.EXPECT().Requested().AnyTimes()
	m.EXPECT().Response(gomock.Any()).AnyTimes()
	h.EXPECT().Do(gomock.Any()).Return(&http.Response{StatusCode: 200}, nil).AnyTimes()

	c.EXPECT().Start(gomock.Any(), gomock.Any()).DoAndReturn(func(done <-chan struct{}, buffer int) chan commander.Event {
		return make(chan commander.Event)
	}).Times(2)
	p.EXPECT().Process(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(done chan struct{}, work <-chan commander.Event, host string, buffer int) chan *http.Request {
			ch := make(chan *http.Request, buffer)
			go func() {
				for {
					r, _ := http.NewRequest(http.MethodGet, host, nil)
					select {
					case ch <- r:
					case <-done:
						return
					}
				}
			}()
			return ch
		}).Times(2)

	cmdr := commander.NewCommander(c, p, h, l, m)
	v := cmdr.Start("http://localhost:9999", 2, 4, 8, 8)

	time.Sleep(200 * time.Millisecond)
	cmdr.Stop(v.ID)
	time.Sleep(100 * time.Millisecond)

	got := cmdr.Vortexes()
	if len(got) != 1 || !got[0].Finished() {
		t.Errorf("expected one finished run, got %+v", got)
	}
}
