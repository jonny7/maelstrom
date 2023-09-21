package domain

import "testing"

func TestWorkerAlreadyExists(t *testing.T) {
	const entry = "0.0.0.0:8080"
	workers := Workers{Workers: []worker{{entry}}}

	if ok := workers.Contains(entry); !ok {
		t.Errorf("expected true, but got %t", ok)
	}
}
