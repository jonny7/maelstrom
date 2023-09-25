package worker

//go:generate mockgen -source=repository.go -destination mock_store.go -package worker
type Store interface {
	Add(worker Worker) error
	Remove(worker Worker) error
	List() []Worker
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{
		store: store,
	}
}

func (s Service) AddWorker(worker string) error {
	return s.store.Add(Worker{IP: worker})
}

func (s Service) RemoveWorker(worker string) error {
	return s.store.Remove(Worker{IP: worker})
}

func (s Service) List() []Worker {
	return s.store.List()
}
