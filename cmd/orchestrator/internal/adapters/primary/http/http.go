package http

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application/config"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application/gateway"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	app        gateway.Orchestrator
	Logger     zerolog.Logger
	HttpServer *http.Server
}

func New(app application.Application, cfg config.Config) Server {
	logger := zerolog.New(os.Stdout).With().Str("service", "maelstrom orchestrator").Timestamp().Logger()
	srv := Server{app: app}
	srv.HttpServer = &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), Handler: srv.routes()}
	srv.Logger = logger
	return srv
}

func (s Server) Run() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	errs := make(chan error)
	go func() {
		s.Logger.Info().Msgf("starting server on: %s", s.HttpServer.Addr)
		errs <- s.HttpServer.ListenAndServe()
		close(errs)
	}()

	select {
	case signals := <-sigs:
		s.Logger.Info().Msg(fmt.Sprintf("shutting down from signal: %v", signals))
		if err := s.HttpServer.Shutdown(context.Background()); err != nil {
			return err
		}
	case err := <-errs:
		s.Logger.Fatal().Msg(fmt.Sprintf("returning from ListenAndServe: %v", err))
		return err
	}
	return nil
}

func (s Server) routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", s.HealthHandler())
	return r
}

func (s Server) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Info().Msgf("%t", s.app.HealthHandler())
		w.Write([]byte("yes"))
		w.WriteHeader(204)
	}
}
