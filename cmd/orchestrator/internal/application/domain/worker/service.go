package worker

type Service struct {
	store Store
}

func NewService(storeType SubscriberStore) (*Service, error) {
	store, err := newSubscriberStore(storeType)
	if err != nil {
		return nil, err
	}
	return newService(store), nil
}

func newService(store Store) *Service {
	return &Service{store: store}
}

func (s Service) Add(remoteAddr string) error {
	return s.store.Add(remoteAddr)
}

func (s Service) Remove(remoteAddr string) error {
	return s.store.Remove(remoteAddr)
}

func (s Service) List() []string {
	return s.store.List()
}
