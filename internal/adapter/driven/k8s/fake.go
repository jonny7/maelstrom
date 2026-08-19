package k8s

// Fake satisfies application.Scaler for developing or running Maelstrom outside k8s.
type Fake struct{}

func NewFake() Fake {
	return Fake{}
}

func (f Fake) Scale(replicas int) error {
	return nil
}
