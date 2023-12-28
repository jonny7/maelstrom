package kafka

import (
	"fmt"
	"time"
)

type Kafka struct{}

func (k Kafka) Consume() {
	for {
		time.Sleep(1 * time.Second)
		fmt.Println("consuming and doing something")
	}
}
