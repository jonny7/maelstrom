package config

type mechanism string

const (
	kafka = "kafka"
)

type Config struct {
	Type mechanism
}
