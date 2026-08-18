package http

import (
	"net/http"
	"slices"
	"testing"
)

func TestDo(t *testing.T) {
	cl := NewFakeHTTP()
	req, _ := http.NewRequest(http.MethodGet, "https://localhost:8080", nil)
	response, _ := cl.Do(req)
	if !slices.Contains(statuses, response.StatusCode) {
		t.Errorf("expected a pre-defined status code of %v, but got %d", statuses, response.StatusCode)
	}
}
