package commander

import (
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

type Processor interface {
	Process(work <-chan Event) chan *http.Request
}

type Consumer interface {
	Consume() chan Event
	Close()
}

// Leave returns the attempted raft removal of the node
func (c *Commander) Leave(id string) error {
	return c.raft.RemoveServer(raft.ServerID(id), 0, 0).Error()
}

// Join returns the attempt to add a new voter to the cluster
func (c *Commander) Join(id, addr string) error {
	return nil
}

func (c *Commander) Consume() chan Event {
	return c.consumer.Consume()
}

func (c *Commander) Process(work chan Event) {
	// @todo add done chans
	ch := c.processor.Process(work)
	go func() {
		defer close(ch)
		for request := range ch {
			response, err := c.client.Do(request)
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}
			log.Info().Msgf("response code was: %d", response.StatusCode)
		}
	}()
}

func NewCommander(cfg Config, consumer Consumer, processor Processor, client HTTPClient) Commander {
	cmdr := Commander{
		config:    cfg,
		consumer:  consumer,
		processor: processor,
		client:    client,
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
