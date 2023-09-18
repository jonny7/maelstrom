package worker

// SubscriberStore is the available type of subscribing stores
type SubscriberStore string

const (
	InMemory SubscriberStore = "in-memory"
)

type worker struct {
	ip string
}

type Workers struct {
	Workers []worker
}
