package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port int    `env:"PORT" envDefault:"4000"`
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	// PrometheusAddr is where the UI reads chart metrics from
	PrometheusAddr string `env:"PROMETHEUS_ADDR" envDefault:"http://localhost:9090"`
	PrometheusUser string `env:"PROMETHEUS_USER" envDefault:"admin"`
	PrometheusPass string `env:"PROMETHEUS_PASS" envDefault:"admin"`
}

func New() (*Config, error) {
	var cfg Config
	opts := env.Options{Prefix: "UI_"}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return nil, fmt.Errorf("UI config failed to load: %w", err)
	}
	return &cfg, nil
}
