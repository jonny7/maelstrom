package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	healthHandler().ServeHTTP(recorder, req)

	want := 204
	got := recorder.Code

	if got != want {
		t.Errorf("expected %d, but got: %d", want, got)
	}
}
