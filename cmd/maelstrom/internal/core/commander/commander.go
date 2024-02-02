package commander

import (
	"net/http"
	"sync"

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/requester"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
	"github.com/rs/zerolog/log"
)

//go:generate mockgen -source=commander.go -destination mock_commander.go -package mocks

type Commander struct {
	config Config
	//raft      *raft.Raft
	consumer  Consumer
	processor Processor
	client    HTTPClient
	interrupt chan struct{}
	logger    logging.Logger
	metrics   metrics.Metrics
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Processor deserializes and converts events to http requests
type Processor interface {
	Process(done chan struct{}, work <-chan requester.Event, host string) chan *http.Request
}

// Consumer provides a mechanism to consume events from any source system
type Consumer interface {
	Start(<-chan struct{}) chan requester.Event
	Close()
}

func (c *Commander) Stop() {
	close(c.interrupt)
}

func (c *Commander) Start(host string, workers int) {
	c.interrupt = make(chan struct{})

	work := c.consumer.Start(c.interrupt)
	load := c.processor.Process(c.interrupt, work, host)

	results := make([]<-chan result, workers)
	for i := 0; i < workers; i++ {
		results[i] = c.vortexer(c.interrupt, load)
	}

	merged := merge(c.interrupt, results...)
	c.analytics(c.interrupt, merged)
}

func merge(done chan struct{}, channels ...<-chan result) chan result {
	var wg sync.WaitGroup
	ch := make(chan result)

	multiplex := func(c <-chan result) {
		defer wg.Done()
		for r := range c {
			select {
			case <-done:
				break
			case ch <- r:
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}

func (c *Commander) analytics(done chan struct{}, results chan result) {
	go func() {
		for {
			select {
			case <-done:
				return
			case res := <-results:
				if res.response != nil {
					c.metrics.Response(res.response.StatusCode)
				} else {
					c.metrics.Response(410)
				}
			}
		}
	}()
}

func deriveStatusCode(resp *http.Response) int {
	if resp == nil {
		return 500
	}
	return resp.StatusCode
}

// Leave returns the attempted raft removal of the node
func (c *Commander) Leave(id string) error {
	return nil //c.raft.RemoveServer(raft.ServerID(id), 0, 0).Error()
}

// Join returns the attempt to add a new voter to the cluster
func (c *Commander) Join(id, addr string) error {
	return nil
}

type result struct {
	err      error
	response *http.Response
}

func (c *Commander) vortexer(done chan struct{}, work chan *http.Request) chan result {
	ch := make(chan result, 1_000_000)
	go func() {
		defer close(ch)
		for {
			select {
			case <-done:
				c.logger.Log(logging.InfoLevel, "commander worker received stop signal for http vortex")
				return
			case request := <-work:
				if request == nil {
					c.logger.Log(logging.ErrorLevel, "nil request")
					continue
				}
				go func() {
					c.logger.Log(logging.DebugLevel, "sending HTTP request")
					c.metrics.Requested()
					response, err := c.client.Do(request)
					if err != nil {
						log.Error().Err(err).Send()
					}
					ch <- result{
						err:      err,
						response: response,
					}
				}()
			}
		}
	}()
	return ch
}

func NewCommander(cfg Config, consumer Consumer, processor Processor, client HTTPClient, logger logging.Logger, metrics metrics.Metrics) Commander {
	cmdr := Commander{
		config:    cfg,
		consumer:  consumer,
		processor: processor,
		client:    client,
		interrupt: make(chan struct{}),
		logger:    logger,
		metrics:   metrics,
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
