package agent

import (
	"github.com/google/uuid"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/service/agent/dto"
)

type Service struct {
	agent agent.Service
}

func NewAgentService(a agent.Service) Service {
	return Service{agent: a}
}

func (s Service) Membership() []dto.Member {
	return dto.MemberDTO(s.agent.Members())
}

func (s Service) StartVortex(host string, jobs int, workers int, cbuf int, rbuf int) {
	s.agent.Start(host, jobs, workers, cbuf, rbuf)
}

func (s Service) EndVortex(id uuid.UUID) {
	s.agent.Stop(id)
}

func (s Service) Vortexes() []maelstrom.Vortex {
	return s.agent.Vortexes()
}
