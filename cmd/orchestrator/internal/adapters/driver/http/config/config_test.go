package config

import (
	"fmt"
	"testing"
)

func TestConfigAddress(t *testing.T) {
	const (
		port = 8081
		host = "0.0.0.0"
	)
	cfg := Config{
		Port: port,
		Host: host,
	}
	want := fmt.Sprintf("%s:%d", host, port)
	got := cfg.HttpAddress()

	if got != want {
		t.Errorf("wanted %s, got %s", want, got)
	}
}
