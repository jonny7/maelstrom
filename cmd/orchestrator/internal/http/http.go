package http

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/config"
	"github.com/rs/zerolog"
	"net/http"
	"os"
)

type Server struct {
	Logger     *zerolog.Logger
	HttpServer *http.Server
}

func New(cfg config.Config) Server {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	mux := router()
	srv := &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), Handler: mux}

	return Server{
		Logger:     &logger,
		HttpServer: srv,
	}

}

func router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", healthHandler())
	return r
}

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}
}
