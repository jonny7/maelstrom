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
}

func New() (*Kafka, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kafka: %w", err)
	}
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Seeds...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.ConsumeTopics(cfg.ConsumerTopics...),
	)
	if err != nil {
		return nil, err
	}
	return &Kafka{client: cl}, nil
}

func (k Kafka) Consume() chan commander.Event {
	ch := make(chan commander.Event)
	go func() {
		for {
			fetch := k.client.PollFetches(context.Background())
			if errs := fetch.Errors(); len(errs) > 0 {
				log.Error().Interface("kafka fetch errors", errs).Send()
			}
			iter := fetch.RecordIter()
			for !iter.Done() {
				record := iter.Next()
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
			}
		}
	}()
	return ch
}

func (k Kafka) Close() {
	log.Info().Msg("closing kafka")
	k.client.Close()
}
