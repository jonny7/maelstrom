package api

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/api/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Server struct {
	app    application.App
	logger zerolog.Logger
	config config.Config
}

func (s Server) Start(w http.ResponseWriter, r *http.Request) {
	s.app.Agent.Start()
	t := true
	render.Respond(w, r, maelstrom.Bool{Success: &t})
}

func New(app application.App) Server {
	logger := zerolog.New(os.Stdout).With().Str("subsystem", "maelstrom API").Timestamp().Logger()
	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	srv := Server{app: app, config: *cfg}
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
	router.Post("/scale", s.Scale)
	router.Post("/start", s.Start)
}

func (s Server) Run(errs chan error, mountRouter func(router chi.Router) http.Handler) {
	// create router
	mux := chi.NewRouter()
	setupMiddlewares(mux)
	// generate routes
	s.setupRoutes(mux)
	// create base router
	base := chi.NewRouter()
	base.Mount("/api", mountRouter(mux))

	go func() {
		s.logger.Info().Msgf("starting API on: %s", s.config.HttpAddress())
		errs <- http.ListenAndServe(s.config.HttpAddress(), base)
	}()
}

func (s Server) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(204)
}

func (s Server) Nodes(w http.ResponseWriter, r *http.Request) {
	members := s.app.Agent.Members()
	render.Respond(w, r, members)
}

func (s Server) Scale(w http.ResponseWriter, r *http.Request) {
	var newScale maelstrom.Scale
	if err := render.Decode(r, &newScale); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.Respond(w, r, maelstrom.Error{Message: "unable to decode body"})
		return
	}

	cfg, buildErr := clientcmd.BuildConfigFromFlags("", "")
	if buildErr != nil {
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, maelstrom.Error{Message: fmt.Sprintf("%e", buildErr)})
	}

	clientSet, cfgErr := kubernetes.NewForConfig(cfg)
	if cfgErr != nil {
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, maelstrom.Error{Message: fmt.Sprintf("%e", cfgErr)})
	}

	cur, getErr := clientSet.AppsV1().
		StatefulSets(s.config.K8s.Namespace).
		GetScale(context.Background(), "maelstrom", metav1.GetOptions{})
	if getErr != nil {
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, maelstrom.Error{Message: fmt.Sprintf("%e", cfgErr)})
	}

	sc := *cur
	sc.Spec.Replicas = int32(*newScale.NumberOfWorkers)

	_, err := clientSet.AppsV1().
		StatefulSets(s.config.K8s.Namespace).
		UpdateScale(context.Background(), "maelstrom", &sc, metav1.UpdateOptions{})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.Respond(w, r, maelstrom.Error{Message: fmt.Sprintf("%e", err)})
	}

	render.Status(r, http.StatusAccepted)
}
