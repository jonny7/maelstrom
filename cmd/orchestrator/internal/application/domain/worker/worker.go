package worker

import (
	"net"
)

// SubscriberStore is the available type of subscribing stores
type SubscriberStore string

const (
	InMemory SubscriberStore = "in-memory"
)

type worker struct {
	ip net.IP
}

type Workers struct {
	Workers []worker
}
