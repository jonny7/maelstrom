package commander

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

type process struct{}

func (p process) Process(work <-chan Event) chan *http.Request {
	ch := make(chan *http.Request)
	done := make(chan struct{})
	go func() {
		time.Sleep(2 * time.Second)
		done <- struct{}{}
	}()
	go func() {
		defer close(ch)
		for {
			select {
			case <-work:
				r, _ := http.NewRequest(http.MethodGet, "localhost:8080", nil)
				ch <- r
			case <-done:
				return
			}
		}
	}()
	return ch
}

type client struct{}

func (c client) Do(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
	}, nil
}

func TestNewCommander(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)

	_ = NewCommander(Config{}, consumer, processor, h)
}

func TestCommanderConsumer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)

	consumer.EXPECT().Consume().Times(1)

	cmdr := NewCommander(Config{}, consumer, processor, h)

	cmdr.consumer.Consume()
}

func TestCommanderProcessor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)

	consumer.EXPECT().Consume().Times(1)

	cmdr := NewCommander(Config{}, consumer, processor, h)

	wrk := cmdr.consumer.Consume()

	processor.EXPECT().Process(wrk).Times(1)

	cmdr.processor.Process(wrk)
}

func TestCommanderProcess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmdr := Commander{
		config:    Config{},
		processor: process{},
		client:    client{},
	}

	ch := make(chan Event)
	go func() {
		defer close(ch)
		for i := 0; i < 5; i++ {
			ch <- Event{
				Key:       []byte(strconv.Itoa(i)),
				Value:     []byte(fmt.Sprintf("message-%d", i)),
				Headers:   nil,
				Timestamp: time.Now(),
			}
		}
	}()
	cmdr.Process(ch)
}
