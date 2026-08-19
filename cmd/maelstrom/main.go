package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jonny7/maelstrom/internal/adapter/driven/k8s"
	"github.com/jonny7/maelstrom/internal/adapter/driven/kafka"
	"github.com/jonny7/maelstrom/internal/adapter/driven/metrics/prometheus"
	"github.com/jonny7/maelstrom/internal/adapter/driven/process"
	"github.com/jonny7/maelstrom/internal/adapter/driven/serf"
	a "github.com/jonny7/maelstrom/internal/adapter/driver/api"
	u "github.com/jonny7/maelstrom/internal/adapter/driver/ui"
	"github.com/jonny7/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/internal/logging"
	"github.com/jonny7/maelstrom/internal/maelstrom"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger, err := logging.NewLogger(logging.InfoLevel, 1, os.Stdout, "service", "Maelstrom Application")
	if err != nil {
		log.Fatal(err)
	}

	// @todo add flag to disable/enable
	m := prometheus.NewMetrics()

	go func() {
		// @todo move this
		http.Handle("/", promhttp.Handler())
		_ = http.ListenAndServe(":2112", nil)
	}()

	// @todo make env or arg
	g, err := kafka.NewFake(m) //kafka.MustNewKafka(logger, m)
	if err != nil {
		log.Fatal(err)
	}

	defer g.Close()

	proc := process.NewProcessor(logger, m)
	if err != nil {
		log.Fatal(err)
	}

	// @todo extend config vars
	h := &http.Client{Timeout: 1500 * time.Millisecond}

	// scale through k8s when running in a cluster, otherwise fall back to the fake
	var scaler application.Scaler
	scaler, err = k8s.New()
	if err != nil {
		logger.LogWithError(logging.WarningLevel, "k8s unavailable, falling back to fake scaler", err)
		scaler = k8s.NewFake()
	}

	var serfCfg serf.Config
	if err = env.Parse(&serfCfg); err != nil {
		return fmt.Errorf("serf config failed to load: %w", err)
	}

	membership, err := serf.New(serfCfg)
	if err != nil {
		return fmt.Errorf("unable to join cluster: %w", err)
	}

	app := application.New(g, proc, h, logger, m, scaler, membership)

	api := a.New(app)

	errs := make(chan error)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})

	api.Run(done, errs, func(router chi.Router) http.Handler {
		return maelstrom.HandlerFromMux(api, router)
	})

	// @todo headless
	ui := u.New(app)
	ui.Run(done, errs)

	for {
		select {
		case msg := <-sig:
			logger.Log(logging.InfoLevel, fmt.Sprintf("shutting down from signal: %v", msg))
			// tell the cluster we're leaving so peers mark us left instead of failed
			if err = app.Leave(); err != nil {
				logger.LogWithError(logging.ErrorLevel, "failed to leave cluster gracefully", err)
			}
			close(done)
			return nil
		case err = <-errs:
			logger.LogWithError(logging.ErrorLevel, "returning from ListenAndServe", err)
			return err
		}
	}
}
