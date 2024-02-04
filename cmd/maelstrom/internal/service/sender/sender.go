package sender

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/sender"
)

type Service struct {
	sender.HTTPDoer
}

func NewSender(s sender.HTTPDoer) Service {
	return Service{s}
}
