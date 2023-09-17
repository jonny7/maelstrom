package http

import (
	"context"
	"encoding/json"
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
	logger     zerolog.Logger
	httpServer *http.Server
}

func New(app application.App, cfg config.Config) Server {
	srv := Server{app: app}
	srv.httpServer = &http.Server{Addr: cfg.HttpAddress(), Handler: srv.routes()}
	srv.logger = zerolog.New(os.Stdout).With().Str("service", "maelstrom orchestrator").Timestamp().Logger()
	return srv
}

func (s Server) Run() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	errs := make(chan error)
	go func() {
		s.logger.Info().Msgf("starting server on: %s", s.httpServer.Addr)
		errs <- s.httpServer.ListenAndServe()
		close(errs)
	}()

	select {
	case signals := <-sigs:
		s.logger.Info().Msgf("shutting down from signal: %v", signals)
		if err := s.httpServer.Shutdown(context.Background()); err != nil {
			return err
		}
	case err := <-errs:
		s.logger.Fatal().Msgf("returning from ListenAndServe: %v", err)
		return err
	}
	return nil
}

func (s Server) routes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", s.healthHandler())
	r.Route("/api/subscribers", func(r chi.Router) {
		r.Post("/", s.createSubscriber)
		r.Get("/", s.listSubscribers)
	})
	return r
}

func (s Server) listSubscribers(w http.ResponseWriter, _ *http.Request) {
	s.logger.Info().Msg("list-subs")

	w.Header().Set("content-type", "application/json")
	b, _ := json.Marshal(s.app.Subscriber.List())
	w.Write(b)
}

func (s Server) createSubscriber(w http.ResponseWriter, r *http.Request) {
	s.logger.Info().Msgf("received subscription from: %s", r.RemoteAddr)
	if err := s.app.Subscriber.Push(r.RemoteAddr); err != nil {
		w.WriteHeader(500)
	}
}

func (s Server) healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}
}

func (s Server) subscribeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Msgf("received subscription from: %s", r.RemoteAddr)
		if err := s.app.Subscriber.Push(r.RemoteAddr); err != nil {
			w.WriteHeader(500)
		}
	}
}
