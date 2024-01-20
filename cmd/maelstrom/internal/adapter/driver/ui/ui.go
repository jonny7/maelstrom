package ui

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jonny7/maelstrom/cmd/maelstrom/dto"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/ui/components"
	"github.com/jonny7/maelstrom/ui/views"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type UI struct {
	app    application.App
	cfg    config.Config
	logger zerolog.Logger
}

func (u UI) Stop(w http.ResponseWriter, _ *http.Request) {
	u.app.Agent.Stop()
	w.WriteHeader(200)
}

func (u UI) Start(w http.ResponseWriter, _ *http.Request) {
	u.app.Agent.Start()
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
	logger := zerolog.New(os.Stdout).With().Str("subsystem", "maelstrom UI").Timestamp().Logger()
	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
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
		u.logger.Info().Msgf("starting UI on: %s", u.cfg.HttpAddress())
		errs <- srv.ListenAndServe()
	}()
	go func() {
		for {
			select {
			case <-done:
				if err := srv.Shutdown(context.Background()); err != nil {
					log.Error().Err(err).Send()
				}
				log.Info().Msg("shutdown UI server... Goodbye!")
				return
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
