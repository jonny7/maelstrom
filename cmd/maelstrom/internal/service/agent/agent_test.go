package agent

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/agent/mocks"
	"go.uber.org/mock/gomock"
)

func setup(t *testing.T) *mocks.MockService {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return mocks.NewMockService(ctrl)
}

func TestMembership(t *testing.T) {
	m := setup(t)
	m.EXPECT().Members()

	svc := NewAgentService(m)
	svc.Membership()
}

func TestVortexes(t *testing.T) {
	m := setup(t)
	m.EXPECT().Vortexes()

	svc := NewAgentService(m)
	svc.Vortexes()
}

func TestStartVortex(t *testing.T) {
	m := setup(t)
	m.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

	svc := NewAgentService(m)
	svc.StartVortex("host.com", 1, 1, 1, 1)
}

func TestEndVortex(t *testing.T) {
	m := setup(t)
	u := uuid.New()
	m.EXPECT().Stop(u)

	svc := NewAgentService(m)
	svc.EndVortex(u)
}
