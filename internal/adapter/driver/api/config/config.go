package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port int    `env:"PORT" envDefault:"3000"`
	Host string `env:"HOST" envDefault:"0.0.0.0"`
}

func New() (*Config, error) {
	var cfg Config
	opts := env.Options{Prefix: "API_"}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return nil, fmt.Errorf("api config failed to load: %w", err)
	}
	return &cfg, nil
}
