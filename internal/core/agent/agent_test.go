package agent

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	mock "github.com/jonny7/maelstrom/internal/core/commander/mocks"
	mocklogger "github.com/jonny7/maelstrom/internal/logging/mocks"
)

// stubMembership records Leave calls and returns a fixed member list.
type stubMembership struct {
	leaves  int
	members []Member
}

func (s *stubMembership) Members() []Member { return s.members }
func (s *stubMembership) Leave() error      { s.leaves++; return nil }

func TestAgentMembers(t *testing.T) {
	s := &stubMembership{members: []Member{{Name: "node-0", Status: "alive"}}}
	a := agent{membership: s}

	got := a.Members()
	if len(got) != 1 || got[0].Name != "node-0" {
		t.Errorf("expected the stub's member list, got %v", got)
	}
}

func TestAgentLeave(t *testing.T) {
	s := &stubMembership{}
	a := agent{membership: s}

	if err := a.Leave(); err != nil {
		t.Errorf("expected no error: %v", err)
	}
	if s.leaves != 1 {
		t.Errorf("expected one Leave call, got %d", s.leaves)
	}
}

func TestAgentStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmdr := mock.NewMockCommand(ctrl)
	l := mocklogger.NewMockLogger(ctrl)
	l.EXPECT().Log(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	a := agent{
		commander: cmdr,
		logger:    l,
	}
	cmdr.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	a.Start("", 1, 1, 0, 0)
}

func TestAgentStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmdr := mock.NewMockCommand(ctrl)
	l := mocklogger.NewMockLogger(ctrl)
	l.EXPECT().Log(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	a := agent{
		commander: cmdr,
		logger:    l,
	}
	u := uuid.New()
	cmdr.EXPECT().Stop(u)
	a.Stop(u)
}
