package main

import (
	_ "embed"
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

	"github.com/jonny7/maelstrom/cmd/maelstrom/common/logging"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/k8s"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/kafka"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/metrics/prometheus"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driven/process"
	a "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/api"
	u "github.com/jonny7/maelstrom/cmd/maelstrom/internal/adapter/driver/ui"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/core/agent"
	"github.com/jonny7/maelstrom/cmd/maelstrom/internal/maelstrom"
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

	var agentCfg agent.Config
	if err = env.Parse(&agentCfg); err != nil {
		return fmt.Errorf("agent config failed to load: %w", err)
	}

	app, err := application.New(agentCfg, g, proc, h, logger, m, scaler)
	if err != nil {
		return err
	}

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
			close(done)
			return nil
		case err = <-errs:
			logger.LogWithError(logging.ErrorLevel, "returning from ListenAndServe", err)
			return err
		}
	}
}
