package ui

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/dto"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui/components/nodes"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui/components/pagination"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui/components/replicas"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui/config"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui/views"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promcfg "github.com/prometheus/common/config"
)

type UI struct {
	app     application.App
	cfg     config.Config
	logger  logging.Logger
	metrics v1.API
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

	// @todo make configurable
	c, err := api.NewClient(api.Config{
		Address:      "http://localhost:9090",
		RoundTripper: promcfg.NewBasicAuthRoundTripper("admin", "admin", "", "", api.DefaultRoundTripper),
	})
	if err != nil {
		log.Fatal(err)
	}

	v1api := v1.NewAPI(c)

	srv := UI{
		app:     app,
		cfg:     *cfg,
		metrics: v1api,
	}
	srv.logger = logger
	return srv
}

func (u UI) Run(done chan struct{}, errs chan error, mountRouter func(router chi.Router) http.Handler) {
	// create router
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	// generate routes
	u.setupRoutes(mux)

	srv := http.Server{
		Addr:    net.JoinHostPort(u.cfg.Host, strconv.Itoa(u.cfg.Port)),
		Handler: mountRouter(mux),
	}

	go func() {
		u.logger.Log(logging.InfoLevel, fmt.Sprintf("starting UI on: %s", net.JoinHostPort(u.cfg.Host, strconv.Itoa(u.cfg.Port))))
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
	router.Get("/paginate", u.Paginate)
	router.Get("/replicas", u.Replicas)
	router.Put("/replicas", u.ScaleReplicas)
}

func (u UI) Index(w http.ResponseWriter, req *http.Request) {
	if err := views.Index().Render(req.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) Nodes(w http.ResponseWriter, req *http.Request) {
	members := dto.MemberDTO(u.app.Agent.Members())
	sort.Slice(members, func(i, j int) bool {
		return members[i].Name < members[j].Name
	})
	if err := nodes.Nodes(members).Render(req.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
func (u UI) Paginate(w http.ResponseWriter, req *http.Request) {
	if err := pagination.Pagination(len(u.app.Agent.Members())).Render(req.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) Replicas(w http.ResponseWriter, r *http.Request) {
	members := u.app.Agent.Members()
	if err := replicas.Replicas(len(members)).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) ScaleReplicas(w http.ResponseWriter, r *http.Request) {
	v, err := strconv.Atoi(r.FormValue("replicas"))
	if err != nil {
		if re := replicas.ReplicasWithError(len(u.app.Agent.Members()), "unable to parse input").Render(r.Context(), w); re != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	scale := maelstrom.Replicas{Replicas: v}
	_, err = u.app.K8s.Scale(scale.Replicas)
	if err != nil {
		if re := replicas.ReplicasWithError(len(u.app.Agent.Members()), err.Error()).Render(r.Context(), w); re != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err = replicas.Replicas(scale.Replicas).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) FindNodeByID(w http.ResponseWriter, r *http.Request, id string) {
	//TODO implement me
	panic("implement me")
}

func (u UI) Stop(w http.ResponseWriter, _ *http.Request) {
	u.app.Agent.Stop()
	w.WriteHeader(200)
}

func (u UI) Vortex(w http.ResponseWriter, _ *http.Request) {
	// @todo
	u.app.Agent.Start("url", 4, 1, 0, 0)
	w.WriteHeader(200)
}

func (u UI) Health(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

//func (u UI) vortex() map[int64]model.SampleValue {
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	tr := v1.Range{
//		Start: time.Now().Add(-5 * time.Minute),
//		End:   time.Now(),
//		Step:  1 * time.Second,
//	}
//	m, warn, e := u.metrics.QueryRange(ctx, "sum(irate(maelstrom_requested{}[5m]))", tr)
//	if e != nil {
//		// @todo
//		log.Println(warn)
//		log.Fatal(e)
//	}
//	mapData := make(map[int64]model.SampleValue)
//
//	for _, val := range m.(model.Matrix)[0].Values {
//		mapData[val.Timestamp.Unix()] = val.Value
//	}
//	return mapData
//}
