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
	"github.com/google/uuid"

	"github.com/jonny7/maelstrom/internal/adapter/driver/ui/components/chart"
	"github.com/jonny7/maelstrom/internal/adapter/driver/ui/components/nodes"
	"github.com/jonny7/maelstrom/internal/adapter/driver/ui/components/replicas"
	"github.com/jonny7/maelstrom/internal/adapter/driver/ui/components/vortex"
	"github.com/jonny7/maelstrom/internal/application/dto"
	"github.com/jonny7/maelstrom/internal/logging"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promcfg "github.com/prometheus/common/config"

	"github.com/jonny7/maelstrom/internal/adapter/driver/ui/config"
	"github.com/jonny7/maelstrom/internal/adapter/driver/ui/views"
	"github.com/jonny7/maelstrom/internal/application"
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

	c, err := api.NewClient(api.Config{
		Address:      cfg.PrometheusAddr,
		RoundTripper: promcfg.NewBasicAuthRoundTripper(cfg.PrometheusUser, promcfg.Secret(cfg.PrometheusPass), "", "", api.DefaultRoundTripper),
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

func (u UI) Run(done chan struct{}, errs chan error) {
	// create router
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	// generate routes
	u.setupRoutes(mux)

	srv := http.Server{
		Addr:    net.JoinHostPort(u.cfg.Host, strconv.Itoa(u.cfg.Port)),
		Handler: mux,
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
	router.Get("/", u.index)
	router.Get("/healthz", u.health)
	router.Get("/nodes", u.nodes)
	router.Get("/replicas", u.replicas)
	router.Put("/replicas", u.scaleReplicas)
	router.Get("/vortex", u.vortexes)
	router.Post("/vortex", u.startVortex)
	router.Delete("/vortex/{id}", u.endVortex)
	router.Get("/vortex-form", u.vortexForm)
	router.Get("/new-vortex", u.newVortex)
	router.Get("/chart", u.chart)
}

func (u UI) index(w http.ResponseWriter, r *http.Request) {
	if err := views.Index().Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) nodes(w http.ResponseWriter, r *http.Request) {
	members := u.app.Membership()
	sort.Slice(members, func(i, j int) bool {
		return members[i].Name < members[j].Name
	})
	if err := nodes.Nodes(members).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// aliveCount ignores members lingering as left or failed in serf's member list.
func aliveCount(members []dto.Member) int {
	n := 0
	for _, m := range members {
		if m.Status == "alive" {
			n++
		}
	}
	return n
}

func (u UI) replicas(w http.ResponseWriter, r *http.Request) {
	members := u.app.Membership()
	if err := replicas.Replicas(aliveCount(members)).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) scaleReplicas(w http.ResponseWriter, r *http.Request) {
	v, err := strconv.Atoi(r.FormValue("replicas"))
	if err != nil {
		if re := replicas.ReplicasWithError(aliveCount(u.app.Membership()), "unable to parse input").Render(r.Context(), w); re != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	err = u.app.Scale(v)
	if err != nil {
		if re := replicas.ReplicasWithError(aliveCount(u.app.Membership()), err.Error()).Render(r.Context(), w); re != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err = replicas.Replicas(v).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) vortexes(w http.ResponseWriter, r *http.Request) {
	runs := u.app.Vortexes()
	if err := vortex.Run(dto.VortexToDTO(runs)).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) startVortex(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(400)
		return
	}

	jobs, err := strconv.ParseInt(r.FormValue("jobs"), 10, 8)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	workers, err := strconv.ParseInt(r.FormValue("workers"), 10, 8)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	cbuf, err := strconv.ParseInt(r.FormValue("consumer_buffer"), 10, 8)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	rbuf, err := strconv.ParseInt(r.FormValue("result_buffer"), 10, 8)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	u.app.StartVortex(r.FormValue("url"), int(jobs), int(workers), int(cbuf), int(rbuf))
	w.Header().Add("HX-Trigger", "updateChartSelectOption")
	if e := vortex.Form().Render(r.Context(), w); e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) endVortex(w http.ResponseWriter, r *http.Request) {
	validatedUUID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(400)
		// @todo
		return
	}
	u.app.EndVortex(validatedUUID)
	w.WriteHeader(202)
}

func (u UI) vortexForm(w http.ResponseWriter, r *http.Request) {
	if err := vortex.Vortex().Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) newVortex(w http.ResponseWriter, r *http.Request) {
	if err := vortex.Form().Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (u UI) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(204)
}

func (u UI) chart(w http.ResponseWriter, r *http.Request) {
	v := u.app.Vortexes()
	sort.Slice(v, func(i, j int) bool {
		return v[i].StartTime.After(v[j].StartTime)
	})
	if err := chart.Chart(dto.VortexToDTO(v)).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
