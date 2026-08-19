package commander

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jonny7/maelstrom/internal/core/metrics"
	"github.com/jonny7/maelstrom/internal/logging"
)

// Vortex is a single load-test run.
type Vortex struct {
	ID             uuid.UUID
	Host           string
	Jobs           int
	Workers        int
	ConsumerBuffer int
	ResultBuffer   int
	StartTime      time.Time
	// EndTime is zero while the run is still going
	EndTime time.Time
}

// Finished reports whether the run has been stopped.
func (v Vortex) Finished() bool {
	return !v.EndTime.IsZero()
}

type Commander struct {
	generator Generator
	processor Processor
	client    HTTPDoer
	logger    logging.Logger
	metrics   metrics.Metrics

	mu   sync.Mutex
	runs map[uuid.UUID]*run
}

// run pairs a Vortex with the channel that stops it
type run struct {
	vortex Vortex
	done   chan struct{}
}

//go:generate mockgen -source=commander.go -destination mocks/commander.go -package mock

type Command interface {
	Vortexes() []Vortex
	Stop(id uuid.UUID)
	Start(host string, jobs int, workers int, cbuf int, rbuf int) Vortex
	Leave(id string) error
	Join(id string, addr string) error
}

// Vortexes returns every run, running or finished.
func (c *Commander) Vortexes() []Vortex {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Vortex, 0, len(c.runs))
	for _, r := range c.runs {
		out = append(out, r.vortex)
	}
	return out
}

// Stop ends the run with the given id. Stopping an unknown or already finished run is a no-op.
func (c *Commander) Stop(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.runs[id]
	if !ok {
		c.logger.LogWithError(logging.ErrorLevel, "vortex run was not found", fmt.Errorf("id: %v was not found", id))
		return
	}
	if r.vortex.Finished() {
		return
	}
	r.vortex.EndTime = time.Now()
	close(r.done)
}

func (c *Commander) Start(host string, jobs, workers, cbuf, rbuf int) Vortex {
	v := Vortex{
		ID:             uuid.New(),
		Host:           host,
		Jobs:           jobs,
		Workers:        workers,
		ConsumerBuffer: cbuf,
		ResultBuffer:   rbuf,
		StartTime:      time.Now(),
	}
	done := make(chan struct{})

	c.mu.Lock()
	c.runs[v.ID] = &run{vortex: v, done: done}
	c.mu.Unlock()

	for range jobs {
		work := c.generator.Start(done, cbuf)
		load := c.processor.Process(done, work, host, cbuf)

		results := make([]<-chan result, workers)
		for w := range workers {
			results[w] = c.vortexer(done, load, rbuf)
		}

		c.analytics(done, merge(done, results...))
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
				return
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

// Leave removes the named node from the cluster. No-op until cluster-wide coordination is implemented.
func (c *Commander) Leave(id string) error {
	return nil
}

// Join adds the named node to the cluster. No-op until cluster-wide coordination is implemented.
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
					c.logger.LogWithError(logging.ErrorLevel, "vortex request failed", err)
				}

				select {
				case ch <- result{err: err, response: response}:
				case <-done:
					return
				}
			}
		}
	}()
	return ch
}

func NewCommander(generator Generator, processor Processor, client HTTPDoer, logger logging.Logger, metrics metrics.Metrics) *Commander {
	return &Commander{
		generator: generator,
		processor: processor,
		client:    client,
		logger:    logger,
		metrics:   metrics,
		runs:      make(map[uuid.UUID]*run),
	}
}
