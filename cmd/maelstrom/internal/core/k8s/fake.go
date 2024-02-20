package k8s

type fakeK8s struct{}

func (f fakeK8s) Scale(replicas int) (int, error) {
	return 0, nil
}

// NewFakeK8s returns a fake k8 client for the purpose of developing or using Maelstrom outside k8s
func NewFakeK8s() K8s {
	return fakeK8s{}
}
