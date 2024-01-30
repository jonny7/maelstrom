package kafka

import (
	"context"
	"fmt"

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/kafka/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Kafka struct {
	client *kgo.Client
	logger logging.Logger
}

func MustNewKafka(logger logging.Logger) (*Kafka, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka: %w", err)
	}
	opts := cfg.DefaultClient()
	opts = append(opts, cfg.WithTLS()...)

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("kakfa client couldn't be initialized: %w", err)
	}

	if err = cl.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka brokers: %w", err)
	}

	if err != nil {
		return nil, err
	}
	return &Kafka{client: cl, logger: logger}, nil
}

func (k Kafka) Start(done <-chan struct{}) chan commander.Event {
	ch := make(chan commander.Event)
	records := k.consume(done)
	go func() {
		defer close(ch)
		for {
			select {
			case record := <-records:
				k.logger.Log(logging.DebugLevel, "")
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
				k.logger.Log(logging.InfoLevel, "exiting kafka consumer")
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
				k.logger.Log(logging.InfoLevel, "closing client consumer")
				return
			default:
				iter := fetches.RecordIter()
				if !iter.Done() {
					record := iter.Next()
					k.logger.Log(logging.DebugLevel, "received message from iterator")
					ch <- record
				}
			}
		}
	}()

	return ch
}

func (k Kafka) Close() {
	k.logger.Log(logging.InfoLevel, "closing kafka")
	k.client.Close()
}
