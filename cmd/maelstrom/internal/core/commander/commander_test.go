package commander

import (
	"net/http"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func newConfig(t *testing.T) Config {
	t.Helper()
	return Config{}
}

func TestNewCommander(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)
	m := NewMockMetrics(ctrl)

	_ = NewCommander(newConfig(t), consumer, processor, h, m)
}

func TestCommanderStop(t *testing.T) {
	done := make(chan struct{})
	cmdr := Commander{
		interrupt: done,
	}
	cmdr.Stop()
	if _, ok := <-done; ok {
		t.Errorf("expected closed channel, but got: %v", done)
	}
}

func TestCommanderStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)
	m := NewMockMetrics(ctrl)

	consumer.EXPECT().Start(gomock.Any()).Times(1)
	processor.EXPECT().Process(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	cmdr := NewCommander(newConfig(t), consumer, processor, h, m)
	cmdr.Start("http://localhost:8000")
}

func TestDeriveStatusCode(t *testing.T) {
	data := []struct {
		name     string
		response *http.Response
		want     int
	}{
		{name: "nil HTTP response", response: nil, want: 500},
		{name: "200 response", response: &http.Response{StatusCode: 200}, want: 200},
		{name: "404 response", response: &http.Response{StatusCode: 404}, want: 404},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			got := deriveStatusCode(d.response)
			if got != d.want {
				t.Errorf("expected %d but got %d", d.want, got)
			}
		})
	}
}

func TestAnalytics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)
	m := NewMockMetrics(ctrl)

	m.EXPECT().Increment().MinTimes(1)

	done := make(chan struct{})
	results := make(chan result)

	cmdr := NewCommander(newConfig(t), consumer, processor, h, m)

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

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)
	m := NewMockMetrics(ctrl)

	h.EXPECT().Do(gomock.Any()).MinTimes(1)

	done := make(chan struct{})
	work := make(chan *http.Request)

	cmdr := NewCommander(newConfig(t), consumer, processor, h, m)

	go func() {
		for i := 0; i < 10; i++ {
			r, _ := http.NewRequest(http.MethodGet, "localhost:8080", nil)
			work <- r
		}
	}()

	cmdr.vortexer(done, work)
	time.Sleep(500 * time.Millisecond)
	close(done)
	time.Sleep(500 * time.Millisecond)
}
