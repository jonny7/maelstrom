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

func (p process) Process(done chan struct{}, work <-chan Event) chan *http.Request {
	ch := make(chan *http.Request)
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

func TestCommanderStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)

	done := make(chan struct{})
	consumer.EXPECT().Start(done).Times(1)

	cmdr := NewCommander(Config{}, consumer, processor, h)

	cmdr.Start()
}

func TestCommanderProcessor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := NewMockConsumer(ctrl)
	processor := NewMockProcessor(ctrl)
	h := NewMockHTTPClient(ctrl)

	done := make(chan struct{})
	consumer.EXPECT().Start(done).Times(1)

	cmdr := NewCommander(Config{}, consumer, processor, h)

	wrk := cmdr.consumer.Start(done)

	processor.EXPECT().Process(done, wrk).Times(1)

	cmdr.processor.Process(done, wrk)
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
	fmt.Println(cmdr)
	//done := make(chan struct{})
	//cmdr.processor.Process(done, ch)
}
