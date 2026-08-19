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

// stubCluster records what it was asked to do.
type stubCluster struct {
	scaled  int
	deleted string
	err     error
}

func (s *stubCluster) Scale(replicas int) error {
	s.scaled = replicas
	return s.err
}

func (s *stubCluster) DeleteNode(node string) error {
	s.deleted = node
	return s.err
}

func TestScale(t *testing.T) {
	s := &stubCluster{}
	app := App{cluster: s}

	if err := app.Scale(3); err != nil {
		t.Errorf("expected no error: %v", err)
	}
	if s.scaled != 3 {
		t.Errorf("expected cluster to receive 3, got %d", s.scaled)
	}
}

func TestScaleInvalidReplicas(t *testing.T) {
	s := &stubCluster{scaled: -99}
	app := App{cluster: s}

	if err := app.Scale(-1); !errors.Is(err, ErrInvalidReplicas) {
		t.Errorf("expected ErrInvalidReplicas, got: %v", err)
	}
	if s.scaled != -99 {
		t.Error("cluster must not be called for invalid input")
	}
}

func TestDeleteNode(t *testing.T) {
	s := &stubCluster{}
	app := App{cluster: s}

	if err := app.DeleteNode("maelstrom-1"); err != nil {
		t.Errorf("expected no error: %v", err)
	}
	if s.deleted != "maelstrom-1" {
		t.Errorf("expected cluster to receive maelstrom-1, got %q", s.deleted)
	}
}

func TestDeleteNodeMissingName(t *testing.T) {
	s := &stubCluster{}
	app := App{cluster: s}

	if err := app.DeleteNode(""); !errors.Is(err, ErrMissingNode) {
		t.Errorf("expected ErrMissingNode, got: %v", err)
	}
	if s.deleted != "" {
		t.Error("cluster must not be called without a node name")
	}
}
