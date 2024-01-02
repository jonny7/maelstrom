package driven

type Nop struct{}

func (n Nop) Consume() {}

func (n Nop) Close() {}
