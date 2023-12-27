package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port int    `env:"PORT" envDefault:"4000"`
	Host string `env:"HOST" envDefault:"0.0.0.0"`
}

func (c Config) HttpAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func New() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config failed to load: %w", err)
	}
	return &cfg, nil
}
