package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/kafka/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/requester"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/metrics"
)

type Kafka struct {
	client  *kgo.Client
	logger  logging.Logger
	metrics metrics.Metrics
}

func MustNewKafka(logger logging.Logger, metrics metrics.Metrics) (*Kafka, error) {
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
	return &Kafka{
		client:  cl,
		logger:  logger,
		metrics: metrics,
	}, nil
}

func (k Kafka) Start(done <-chan struct{}, buffer int) chan requester.Event {
	ch := make(chan requester.Event, buffer)
	records := k.consume(done)
	go func() {
		defer close(ch)
		for {
			select {
			case record := <-records:
				event := requester.Event{
					Key:   record.Key,
					Value: record.Value,
					Headers: func(record *kgo.Record) []requester.Header {
						var headers []requester.Header
						for _, v := range record.Headers {
							headers = append(headers, requester.Header{
								Key:   v.Key,
								Value: v.Value,
							})
						}
						return headers
					}(record),
					Timestamp: record.Timestamp,
				}
				select {
				case ch <- event:
					k.metrics.Consumed()
				case <-done:
					k.logger.Log(logging.InfoLevel, "exiting kafka consumer")
					return
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

	go func() {
		defer close(ch)
		for {
			select {
			case <-done:
				k.logger.Log(logging.InfoLevel, "closing client consumer")
				return
			default:
				fetches := k.client.PollFetches(context.Background())
				if errs := fetches.Errors(); len(errs) > 0 {
					for _, e := range errs {
						k.logger.LogWithError(logging.ErrorLevel, "Kafka fetch errors", e.Err)
					}
				}
				iter := fetches.RecordIter()
				if !iter.Done() {
					record := iter.Next()
					k.logger.Log(logging.DebugLevel, "received message from iterator")
					select {
					case ch <- record:
					case <-done:
						k.logger.Log(logging.InfoLevel, "closing client consumer")
						return
					}
				}
				if err := k.client.CommitUncommittedOffsets(context.Background()); err != nil {
					k.logger.LogWithError(logging.ErrorLevel, "failed to commit offsets", err)
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
