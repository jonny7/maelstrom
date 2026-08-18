package logging

import (
	"errors"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rzajac/zltest"
)

func TestNewLogger(t *testing.T) {
	data := []struct {
		name string
		rate uint32
		want error
	}{
		{name: "invalid rate", rate: 0, want: invalidRate},
		{name: "valid rate", rate: 10, want: nil},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			_, err := NewLogger(InfoLevel, d.rate, os.Stdout, "serviceKey", "service")
			if !errors.Is(err, d.want) {
				t.Errorf("expected %v, but got: %v", d.want, err)
			}
		})
	}
}

func TestLogLevelMet(t *testing.T) {
	tst := zltest.New(t)

	log, _ := NewLogger(InfoLevel, 1, tst, "serviceKey", "service")

	log.Log(InfoLevel, "test", KV{
		Key:   "key",
		Value: "val",
	})

	ent := tst.LastEntry()
	ent.ExpLevel(zerolog.InfoLevel)
	ent.ExpMsg("test")
	ent.ExpStr("key", "val")
}

func TestLogLevelNotMet(t *testing.T) {
	tst := zltest.New(t)

	log, _ := NewLogger(InfoLevel, 1, tst, "serviceKey", "service")

	log.Log(DebugLevel, "test", KV{
		Key:   "key",
		Value: "val",
	})

	entry := tst.LastEntry()

	if entry != nil {
		t.Errorf("last entry should be nil, but got :v")
	}
}

func TestLogLevelWithErrMet(t *testing.T) {
	tst := zltest.New(t)

	e := errors.New("some error")

	log, _ := NewLogger(WarningLevel, 1, tst, "serviceKey", "service")

	log.LogWithError(ErrorLevel, "test", e, KV{
		Key:   "key",
		Value: "val",
	})

	ent := tst.LastEntry()
	ent.ExpLevel(zerolog.ErrorLevel)
	ent.ExpMsg("test")
	ent.ExpStr("key", "val")
	ent.ExpErr(e)
}

func TestLogLevelWithErrNotMet(t *testing.T) {
	tst := zltest.New(t)

	e := errors.New("some error")

	log, _ := NewLogger(ErrorLevel, 1, tst, "serviceKey", "service")

	log.LogWithError(WarningLevel, "test", e, KV{
		Key:   "key",
		Value: "val",
	})

	entry := tst.LastEntry()

	if entry != nil {
		t.Errorf("last entry should be nil, but got :v")
	}
}
