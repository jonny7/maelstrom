package domain

type WorkerStore interface {
	Add(worker string) error
	Remove(worker string) error
	List() []string
}
