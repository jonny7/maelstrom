package kafka

import (
	"context"
	"fmt"

	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/kafka/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Kafka struct {
	client *kgo.Client
	cfg    config.Config
}

func MustNewKafka() (*Kafka, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka: %w", err)
	}
	opts := cfg.DefaultClient()
	opts = append(opts, cfg.WithTLS()...)

	cl, err := kgo.NewClient(opts...)
	if err = cl.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka brokers: %w", err)
	}

	if err != nil {
		return nil, err
	}
	return &Kafka{client: cl}, nil
}

func (k Kafka) Start(done <-chan struct{}) chan commander.Event {
	ch := make(chan commander.Event)
	records := k.consume(done)
	go func() {
		defer close(ch)
		for {
			select {
			case record := <-records:
				log.Debug().Msg("creating event from franz record")
				ch <- commander.Event{
					Key:   record.Key,
					Value: record.Value,
					Headers: func(record *kgo.Record) []commander.Header {
						var headers []commander.Header
						for _, v := range record.Headers {
							headers = append(headers, commander.Header{
								Key:   v.Key,
								Value: v.Value,
							})
						}
						return headers
					}(record),
					Timestamp: record.Timestamp,
				}
			case <-done:
				log.Info().Msg("exiting kafka consumer")
				return
			}
		}
	}()
	return ch
}

func (k Kafka) consume(done <-chan struct{}) chan *kgo.Record {
	ch := make(chan *kgo.Record)

	fetches := k.client.PollFetches(context.Background())
	if errs := fetches.Errors(); len(errs) > 0 {
		// All errors are retried internally when fetching, but non-retriable errors are
		// returned from polls so that users can notice and take action.
		panic(fmt.Sprint(errs))
	}

	go func() {
		defer close(ch)
		for {
			select {
			case <-done:
				log.Info().Msg("closing client consumer")
				return
			default:
				iter := fetches.RecordIter()
				if !iter.Done() {
					record := iter.Next()
					log.Debug().Msg("received message from iterator")
					ch <- record
					//k.client.CommittedOffsets()
				}
			}
		}
	}()

	return ch
}

func (k Kafka) Close() {
	log.Info().Msg("closing kafka")
	k.client.Close()
}
