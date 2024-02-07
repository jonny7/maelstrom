package ui

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/dto"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/ui/components"
	"github.com/jonny7/maelstrom/ui/views"
)

type UI struct {
	app    application.App
	cfg    config.Config
	logger logging.Logger
}

func (u UI) Stop(w http.ResponseWriter, _ *http.Request) {
	u.app.Agent.Stop()
	w.WriteHeader(200)
}

func (u UI) Start(w http.ResponseWriter, _ *http.Request) {
	// @todo
	u.app.Agent.Start("url", 4, 0, 0)
	w.WriteHeader(200)
}

func (u UI) Health(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (u UI) Scale(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func New(app application.App) UI {
	logger, err := logging.NewLogger(logging.InfoLevel, 1, os.Stdout, "subsystem", "Maelstrom UI")
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	srv := UI{app: app, cfg: *cfg}
	srv.logger = logger
	return srv
}

func (u UI) Run(done chan struct{}, errs chan error, mountRouter func(router chi.Router) http.Handler) {
	// create router
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	// generate routes
	u.setupRoutes(mux)

	srv := http.Server{Addr: u.cfg.HttpAddress(), Handler: mountRouter(mux)}

	go func() {
		u.logger.Log(logging.InfoLevel, fmt.Sprintf("starting UI on: %s", u.cfg.HttpAddress()))
		errs <- srv.ListenAndServe()
	}()
	go func() {
		defer u.logger.Log(logging.InfoLevel, "shutdown UI server... Goodbye!")
		for range done {
			if err := srv.Shutdown(context.Background()); err != nil {
				u.logger.LogWithError(logging.ErrorLevel, "unable to shutdown server", err)
			}
		}
	}()
}

func (u UI) setupRoutes(router *chi.Mux) {
	router.Get("/", u.Index)
	router.Get("/nodes", u.Nodes)
}

func (u UI) Index(w http.ResponseWriter, req *http.Request) {
	if err := views.Index().Render(req.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) Nodes(w http.ResponseWriter, req *http.Request) {
	nodes := dto.MemberDTO(u.app.Agent.Members())
	if err := components.Nodes(nodes).Render(req.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
