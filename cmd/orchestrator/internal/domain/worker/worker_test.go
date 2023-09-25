package worker

import "testing"

func TestWorkerExists(t *testing.T) {
	const entry = "0.0.0.0:8080"
	workers := Workers{Workers: []Worker{{entry}}}

	data := []struct {
		name    string
		workers Workers
		entry   string
		result  bool
	}{
		{name: "worker exists", workers: workers, entry: entry, result: true},
		{name: "worker doesn't exists", workers: workers, entry: "0.0.0.0:8081", result: false},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			if d.result != workers.Contains(d.entry) {
				t.Errorf("expected %t, but got %t", d.result, workers.Contains(d.entry))
			}
		})
	}
}
