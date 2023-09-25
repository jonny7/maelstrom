package worker

// SubscriberStore is the available type of subscribing stores
type SubscriberStore string

type Workers struct {
	Workers []Worker
}

type Worker struct {
	IP string
}

func (w Workers) Contains(ip string) bool {
	for _, work := range w.Workers {
		if ip == work.IP {
			return true
		}
	}
	return false
}
