package http

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/adapters/driver/http/config"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/application"
	"github.com/rs/zerolog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	app        application.App
	Logger     zerolog.Logger
	HttpServer *http.Server
}

func New(app application.App, cfg config.Config) Server {
	srv := Server{app: app}
	srv.HttpServer = &http.Server{Addr: cfg.HttpAddress(), Handler: srv.routes()}
	srv.Logger = zerolog.New(os.Stdout).With().Str("service", "maelstrom orchestrator").Timestamp().Logger()
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
	r.Get("/health", s.healthHandler())
	return r
}

func (s Server) healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}
}
