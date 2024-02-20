package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/goccy/go-json"
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/api/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
)

type Server struct {
	app    application.App
	logger logging.Logger
	cfg    config.Config
}

func (s Server) ScaleReplicas(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s Server) FindNodeByID(w http.ResponseWriter, r *http.Request, id string) {
	//TODO implement me
	panic("implement me")
}

func (s Server) Stop(w http.ResponseWriter, r *http.Request) {
	s.app.Agent.Stop()
	render.Respond(w, r, maelstrom.Status{
		Message: "success",
		Status:  http.StatusOK,
	})
}

func (s Server) Vortex(w http.ResponseWriter, r *http.Request) {
	var vortex maelstrom.Vortex
	b, err := io.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, maelstrom.Status{
			Message: fmt.Sprintf("unable to read body of request: %v", err),
			Status:  http.StatusInternalServerError,
		})
		return
	}
	err = json.Unmarshal(b, &vortex)
	if err != nil {
		render.Status(r, 400)
		render.Respond(w, r, maelstrom.Status{
			Message: fmt.Sprintf("unable to unmarshall request: %v", err),
			Status:  http.StatusBadRequest,
		})
		return
	}
	if _, err = url.Parse(vortex.Host); err != nil {
		render.Status(r, 400)
		render.Respond(w, r, fmt.Sprintf("the provided host was unable to be parsed: %s", vortex.Host))
	}
	s.app.Agent.Start(vortex.Host, vortex.Jobs, vortex.Workers, vortex.ConsumerBuffer, vortex.ResultBuffer)
	render.Respond(w, r, maelstrom.Status{
		Message: "success",
		Status:  http.StatusOK,
	})
}

func New(app application.App) Server {
	logger, err := logging.NewLogger(logging.InfoLevel, 1, os.Stdout, "subsystem", "Maelstrom API")
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	srv := Server{app: app, cfg: *cfg}
	srv.logger = logger
	return srv
}

// setupMiddlewares applies various middlewares to chi
func setupMiddlewares(router *chi.Mux) {
	router.Use(middleware.Recoverer)
	router.Use(render.SetContentType(render.ContentTypeJSON))
}

func (s Server) setupRoutes(router *chi.Mux) {
	router.Get("/healthz", s.Health)
	router.Get("/nodes", s.Nodes)
	router.Post("/replicas", s.Replicas)
	router.Post("/vortex", s.Vortex)
}

func (s Server) Run(done chan struct{}, errs chan error, mountRouter func(router chi.Router) http.Handler) {
	// create router
	mux := chi.NewRouter()
	setupMiddlewares(mux)
	// generate routes
	s.setupRoutes(mux)
	// create base router
	base := chi.NewRouter()
	base.Mount("/api", mountRouter(mux))

	srv := http.Server{Addr: net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port)), Handler: base}

	go func() {
		s.logger.Log(logging.InfoLevel, fmt.Sprintf("starting API on %s", net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))))
		errs <- srv.ListenAndServe()
	}()
	go func() {
		defer s.logger.Log(logging.InfoLevel, "shutdown API server... Goodbye!")
		for range done {
			if err := srv.Shutdown(context.Background()); err != nil {
				s.logger.LogWithError(logging.ErrorLevel, "unable to shutdown server", err)
			}
		}
	}()
}

func (s Server) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(204)
}

func (s Server) Nodes(w http.ResponseWriter, r *http.Request) {
	members := s.app.Agent.Members()
	render.Respond(w, r, members)
}

func (s Server) Replicas(w http.ResponseWriter, r *http.Request) {
	var newScale maelstrom.Replicas
	if err := render.Decode(r, &newScale); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, maelstrom.Status{Message: "unable to decode body"})
		return
	}
	if status, err := s.app.K8s.Scale(newScale.Replicas); err != nil {
		render.Status(r, status)
		render.Respond(w, r, maelstrom.Status{Message: err.Error()})
	}

	render.Status(r, http.StatusAccepted)
}
