package domain

// SubscriberStore is the available type of subscribing stores
type SubscriberStore string

type Workers struct {
	Workers []worker
}

type worker struct {
	ip string
}

func (w Workers) Contains(ip string) bool {
	for _, work := range w.Workers {
		if ip == work.ip {
			return true
		}
	}
	return false
}
