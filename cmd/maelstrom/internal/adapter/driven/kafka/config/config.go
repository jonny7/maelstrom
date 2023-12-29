package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Seeds          []string `env:"BROKERS" envDefault:"localhost:9092"`
	ConsumerGroup  string   `env:"CONSUMER_GROUP" envDefault:"maelstrom"`
	ConsumerTopics []string `env:"CONSUMER_TOPICS,required"`
}

func New() (*Config, error) {
	var cfg Config
	opts := env.Options{Prefix: "KAFKA_"}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return nil, fmt.Errorf("kafka config failed to load: %w", err)
	}
	return &cfg, nil
}
