package processor

import (
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/commander/processor"
)

type Service struct {
	processor.Processor
}

func NewProcessor(p processor.Processor) Service {
	return Service{p}
}
