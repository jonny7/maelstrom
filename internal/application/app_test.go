package application

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	agentmocks "github.com/jonny7/maelstrom/internal/core/agent/mocks"
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

func TestLeave(t *testing.T) {
	m := setup(t)
	m.EXPECT().Leave()

	app := App{agent: m}
	if err := app.Leave(); err != nil {
		t.Errorf("expected no error: %v", err)
	}
}

// stubScaler records the replica count it was asked for.
type stubScaler struct {
	scaled int
	err    error
}

func (s *stubScaler) Scale(replicas int) error {
	s.scaled = replicas
	return s.err
}

func TestScale(t *testing.T) {
	s := &stubScaler{}
	app := App{scaler: s}

	if err := app.Scale(3); err != nil {
		t.Errorf("expected no error: %v", err)
	}
	if s.scaled != 3 {
		t.Errorf("expected scaler to receive 3, got %d", s.scaled)
	}
}

func TestScaleInvalidReplicas(t *testing.T) {
	s := &stubScaler{scaled: -99}
	app := App{scaler: s}

	if err := app.Scale(-1); !errors.Is(err, ErrInvalidReplicas) {
		t.Errorf("expected ErrInvalidReplicas, got: %v", err)
	}
	if s.scaled != -99 {
		t.Error("scaler must not be called for invalid input")
	}
}
