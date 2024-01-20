package commander

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/raft"
	"github.com/rs/zerolog/log"
)

//go:generate mockgen -source=commander.go -destination mock_commander.go -package commander

type Commander struct {
	config    Config
	raft      *raft.Raft
	consumer  Consumer
	processor Processor
	client    HTTPClient
	interrupt chan struct{}
}

type Event struct {
	Key       []byte
	Value     []byte
	Headers   []Header
	Timestamp time.Time
}

type Header struct {
	Key   string
	Value []byte
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Processor deserializes and converts events to http requests
type Processor interface {
	Process(done chan struct{}, work <-chan Event) chan *http.Request
}

// Consumer provides a mechanism to consume events from any source system
type Consumer interface {
	Start(<-chan struct{}) chan Event
	Close()
}

func (c *Commander) Stop() {
	close(c.interrupt)
}

func (c *Commander) Start() {
	c.interrupt = make(chan struct{})
	// @todo redo tests
	work := c.consumer.Start(c.interrupt)
	load := c.processor.Process(c.interrupt, work)
	results := c.Vortex(c.interrupt, load)

	go func() {
		for r := range results {
			// @todo do metrics here
			fmt.Println(r)
		}
	}()
}

// Leave returns the attempted raft removal of the node
func (c *Commander) Leave(id string) error {
	return c.raft.RemoveServer(raft.ServerID(id), 0, 0).Error()
}

// Join returns the attempt to add a new voter to the cluster
func (c *Commander) Join(id, addr string) error {
	return nil
}

type result struct {
	err      error
	response *http.Response
}

func (c *Commander) Vortex(done chan struct{}, work chan *http.Request) chan result {
	// @todo make this parallelizable
	ch := make(chan result)
	go func() {
		defer close(ch)
		for {
			select {
			case <-done:
				log.Debug().Msg("commander received stop signal for http vortex")
				return
			case request := <-work:
				log.Debug().Msg("sending HTTP request")
				response, err := c.client.Do(request)
				ch <- result{
					err:      err,
					response: response,
				}
			}
		}
	}()
	return ch
}

func NewCommander(cfg Config, consumer Consumer, processor Processor, client HTTPClient) Commander {
	cmdr := Commander{
		config:    cfg,
		consumer:  consumer,
		processor: processor,
		client:    client,
		interrupt: make(chan struct{}),
	}

	return cmdr
}

// @todo do later
//func newRaft() {
//	cfg := raft.DefaultConfig()
//	fmt.Print(cfg)
//	_, err := boltdb.New(boltdb.Options{})
//	if err != nil {
//		return
//	}
//}
