package commander

import (
	"net/http"
	"time"

	"github.com/hashicorp/raft"
	"github.com/rs/zerolog/log"
)

type Commander struct {
	config    Config
	raft      *raft.Raft
	consumer  Consumer
	processor Processor
	client    http.Client
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

func (c *Commander) Process(work chan Event) {
	ch := c.processor.Process(work)

	for request := range ch {
		response, err := c.client.Do(request)
		if err != nil {
			log.Error().Err(err).Send()
			continue
		}
		log.Info().Msgf("response code was: %d", response.StatusCode)
	}
}

func NewCommander(cfg Config, consumer Consumer, processor Processor) (*Commander, error) {
	cmdr := &Commander{
		config:    cfg,
		consumer:  consumer,
		processor: processor,
	}

	work := consumer.Consume()

	cmdr.Process(work)

	return cmdr, nil
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
