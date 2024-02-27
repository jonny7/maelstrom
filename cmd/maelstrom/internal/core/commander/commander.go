package commander

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/generator"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/rs/zerolog/log"
)

type Commander struct {
	config Config
	//raft      *raft.Raft
	generator generator.Generator
	processor processor.Processor
	client    sender.HTTPDoer
	interrupt map[uuid.UUID]vortexWithInterupt
	logger    logging.Logger
	metrics   metrics.Metrics
}

type vortexWithInterupt struct {
	maelstrom.Vortex
	done chan struct{}
}

func toInt(i int) *int {
	return &i
}

func toString(s string) *string {
	return &s
}

func (c *Commander) Vortexes() []maelstrom.Vortex {
	var out []maelstrom.Vortex
	interupts := c.interrupt
	for _, v := range interupts {
		out = append(out, v.Vortex)
	}
	return out
}

func (c *Commander) Stop(id uuid.UUID) {
	i, ok := c.interrupt[id]
	if !ok {
		c.logger.LogWithError(logging.ErrorLevel, "vortex run was not found", fmt.Errorf("id: %v was not found", id))
		return
	}
	close(i.done)
	delete(c.interrupt, id)
}

func (c *Commander) Start(host string, jobs, workers, cbuf, rbuf int) maelstrom.Vortex {
	u := uuid.New()
	if c.interrupt == nil {
		c.interrupt = make(map[uuid.UUID]vortexWithInterupt)
	}
	v := maelstrom.Vortex{
		ConsumerBuffer: cbuf,
		EndTime:        nil,
		Host:           &host,
		Id:             toString(u.String()),
		Jobs:           jobs,
		ResultBuffer:   rbuf,
		StartTime:      toInt(int(time.Now().Unix())),
		Workers:        workers,
	}
	c.interrupt[u] = vortexWithInterupt{
		Vortex: v,
		done:   make(chan struct{}),
	}

	for j := 0; j < jobs; j++ {
		work := c.generator.Start(c.interrupt[u].done, cbuf)
		load := c.processor.Process(c.interrupt[u].done, work, host, cbuf)

		results := make([]<-chan result, workers)
		for w := 0; w < workers; w++ {
			results[w] = c.vortexer(c.interrupt[u].done, load, rbuf)
		}

		merged := merge(c.interrupt[u].done, results...)
		c.analytics(c.interrupt[u].done, merged)
	}
	return v
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
				}
			}
		}
	}()
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

func (c *Commander) vortexer(done chan struct{}, work chan *http.Request, buffer int) chan result {
	ch := make(chan result, buffer)
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
			}
		}
	}()
	return ch
}

func NewCommander(cfg Config, generator generator.Generator, processor processor.Processor, client sender.HTTPDoer, logger logging.Logger, metrics metrics.Metrics) Commander {
	cmdr := Commander{
		config:    cfg,
		generator: generator,
		processor: processor,
		client:    client,
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
