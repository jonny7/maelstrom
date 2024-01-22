package metrics

type Fake struct{}

func (f Fake) Increment() {}

func New() Fake {
	return Fake{}
}
