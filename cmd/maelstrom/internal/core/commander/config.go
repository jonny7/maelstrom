package commander

import "github.com/hashicorp/raft"

type Config struct {
	Raft raft.Config // @todo extend to support env config
	ConsumerConfig
}

type ConsumerConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
}
