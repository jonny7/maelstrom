package kafka

import "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander"

type Fake struct{}

func (f Fake) Start(done <-chan struct{}) chan commander.Event {
	//TODO implement me
	panic("implement me")
}

func (f Fake) Close() {}

func NewFake() Fake {
	return Fake{}
}
