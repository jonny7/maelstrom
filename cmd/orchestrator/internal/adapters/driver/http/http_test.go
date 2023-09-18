package http

import (
	"fmt"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application/services"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	app := services.NewApplication()
	srv := Server{app: app}
	srv.httpServer = &http.Server{Addr: fmt.Sprintf("%s:%d", "0.0.0.0", 3000), Handler: srv.routes()}

	srv.healthHandler().ServeHTTP(recorder, req)

	want := 204
	got := recorder.Code

	if got != want {
		t.Errorf("expected %d, but got: %d", want, got)
	}
}

func TestSubscribe(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/subscriber", nil)

	app := services.NewApplication()
	srv := Server{app: app}
	srv.httpServer = &http.Server{Addr: fmt.Sprintf("%s:%d", "0.0.0.0", 3000), Handler: srv.routes()}

	srv.createSubscriber().ServeHTTP(recorder, req)

	want := 202
	got := recorder.Code

	if got != want {
		t.Errorf("expected %d, but got: %d", want, got)
	}
}

func TestGetSubscribers(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/subscriber", nil)

	app := services.NewApplication()
	srv := Server{app: app}
	srv.httpServer = &http.Server{Addr: fmt.Sprintf("%s:%d", "0.0.0.0", 3000), Handler: srv.routes()}

	srv.listSubscribers().ServeHTTP(recorder, req)

	want := 200
	got := recorder.Code

	if got != want {
		t.Errorf("expected %d, but got: %d", want, got)
	}
}
