package ui

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/ui/views"
	"github.com/rs/zerolog"
)

type UI struct {
	app    application.App
	cfg    config.Config
	logger zerolog.Logger
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

func (u UI) Run(errs chan error) {
	// create router
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	// generate routes
	u.setupRoutes(router)

	go func() {
		u.logger.Info().Msgf("starting UI on: %s", u.cfg.HttpAddress())
		errs <- http.ListenAndServe(u.cfg.HttpAddress(), router)
	}()
}

func (u UI) setupRoutes(router *chi.Mux) {
	router.Get("/", u.Index)
}

func (u UI) Index(w http.ResponseWriter, req *http.Request) {
	//members := u.app.Agent.Members()
	views.Index().Render(req.Context(), w)
}
