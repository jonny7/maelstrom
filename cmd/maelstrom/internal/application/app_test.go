package application

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	agentmocks "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/agent/mocks"
	k8smocks "github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/k8s/mocks"
)

func setup(t *testing.T) *agentmocks.MockService {
	t.Helper()
	return agentmocks.NewMockService(gomock.NewController(t))
}

func TestMembership(t *testing.T) {
	m := setup(t)
	m.EXPECT().Members()

	app := App{agent: m}
	app.Membership()
}

func TestVortexes(t *testing.T) {
	m := setup(t)
	m.EXPECT().Vortexes()

	app := App{agent: m}
	app.Vortexes()
}

func TestStartVortex(t *testing.T) {
	m := setup(t)
	m.EXPECT().Start("host.com", 1, 1, 1, 1)

	app := App{agent: m}
	app.StartVortex("host.com", 1, 1, 1, 1)
}

func TestEndVortex(t *testing.T) {
	m := setup(t)
	u := uuid.New()
	m.EXPECT().Stop(u)

	app := App{agent: m}
	app.EndVortex(u)
}

func TestScale(t *testing.T) {
	k := k8smocks.NewMockK8s(gomock.NewController(t))
	k.EXPECT().Scale(3).Return(0, nil)

	app := App{k8s: k}
	if _, err := app.Scale(3); err != nil {
		t.Errorf("expected no error: %v", err)
	}
}
