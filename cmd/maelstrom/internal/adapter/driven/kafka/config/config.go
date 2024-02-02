package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

type Config struct {
	Seeds          []string `env:"CONSUMER_BROKERS"`
	ConsumerGroup  string   `env:"CONSUMER_GROUP" envDefault:"maelstrom"`
	ConsumerTopics []string `env:"CONSUMER_TOPICS,required"`
	User           string   `env:"USER"`
	Password       string   `env:"PASSWORD"`
	EnableTLS      bool     `env:"ENABLE_TLS" envDefault:"false"`
	EnableMetrics  bool     `env:"ENABLE_METRICS" envDefault:"true"`
}

func New() (*Config, error) {
	var cfg Config
	opts := env.Options{Prefix: "KAFKA_"}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return nil, fmt.Errorf("kafka config failed to load: %w", err)
	}
	return &cfg, nil
}

// @todo expand this
func (c Config) DefaultClient() []kgo.Opt {
	var opts []kgo.Opt
	opts = append(opts, kgo.SeedBrokers(c.Seeds...))
	opts = append(opts, kgo.ConsumerGroup(c.ConsumerGroup))
	opts = append(opts, kgo.ConsumeTopics(c.ConsumerTopics...))
	opts = append(opts, kgo.DisableAutoCommit())

	return opts
}

func (c Config) WithTLS() []kgo.Opt {
	var opts []kgo.Opt
	if c.EnableTLS {
		opts = append(opts, kgo.DialTLS())
		opts = append(opts, kgo.SASL(scram.Auth{
			User: c.User,
			Pass: c.Password,
		}.AsSha512Mechanism()))
	}
	return opts
}
