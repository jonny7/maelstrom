package commander

type Config struct {
	ConsumerConfig
}

type ConsumerConfig struct {
	Brokers []string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
}
