package api

import (
	"context"
	"errors"
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
	"github.com/google/uuid"

	"github.com/jonny7/maelstrom/internal/adapter/driver/api/config"
	"github.com/jonny7/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/internal/core/commander"
	"github.com/jonny7/maelstrom/internal/logging"
	"github.com/jonny7/maelstrom/internal/maelstrom"
)

type Server struct {
	app    application.App
	logger logging.Logger
	cfg    config.Config
}

func (s Server) Vortexes(w http.ResponseWriter, r *http.Request) {
	runs := s.app.Vortexes()
	out := make([]maelstrom.Vortex, 0, len(runs))
	for _, v := range runs {
		out = append(out, toAPIVortex(v))
	}
	render.Respond(w, r, out)
}

// toAPIVortex maps a core run onto the generated OpenAPI shape
func toAPIVortex(v commander.Vortex) maelstrom.Vortex {
	id := v.ID.String()
	host := v.Host
	start := int(v.StartTime.Unix())
	out := maelstrom.Vortex{
		Id:             &id,
		Host:           &host,
		Jobs:           v.Jobs,
		Workers:        v.Workers,
		ConsumerBuffer: v.ConsumerBuffer,
		ResultBuffer:   v.ResultBuffer,
		StartTime:      &start,
	}
	if v.Finished() {
		end := int(v.EndTime.Unix())
		out.EndTime = &end
	}
	return out
}

func (s Server) ScaleReplicas(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s Server) FindNodeByID(w http.ResponseWriter, r *http.Request, id string) {
	//TODO implement me
	panic("implement me")
}

func (s Server) EndVortex(w http.ResponseWriter, r *http.Request, id string) {
	validatedUUID, err := uuid.Parse(id)
	if err != nil {
		w.WriteHeader(400)
		// @todo
		return
	}
	s.app.EndVortex(validatedUUID)
	render.Respond(w, r, maelstrom.Status{
		Message: "success",
		Status:  http.StatusOK,
	})
}

func (s Server) StartVortex(w http.ResponseWriter, r *http.Request) {
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
	if _, err = url.Parse(*vortex.Host); err != nil {
		render.Status(r, 400)
		render.Respond(w, r, fmt.Sprintf("the provided host was unable to be parsed: %s", *vortex.Host))
	}
	s.app.StartVortex(*vortex.Host, vortex.Jobs, vortex.Workers, vortex.ConsumerBuffer, vortex.ResultBuffer)
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

func (s Server) Run(done chan struct{}, errs chan error, mountRouter func(router chi.Router) http.Handler) {
	// create router
	mux := chi.NewRouter()
	setupMiddlewares(mux)
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
	members := s.app.Membership()
	render.Respond(w, r, members)
}

func (s Server) Replicas(w http.ResponseWriter, r *http.Request) {
	var newScale maelstrom.Replicas
	if err := render.Decode(r, &newScale); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, maelstrom.Status{Message: "unable to decode body"})
		return
	}
	if err := s.app.Scale(newScale.Replicas); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, application.ErrInvalidReplicas) {
			status = http.StatusUnprocessableEntity
		}
		render.Status(r, status)
		render.Respond(w, r, maelstrom.Status{Message: err.Error(), Status: status})
		return
	}

	render.Status(r, http.StatusAccepted)
	render.Respond(w, r, maelstrom.Status{Message: "success", Status: http.StatusAccepted})
}
