package main

import (
	"context"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/config"
	"github.com/jonny7/maelstrom/cmd/orchestrator/internal/http"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, cfgErr := config.New()
	if cfgErr != nil {
		log.Fatal().Err(cfgErr).Send()
	}

	srv := http.New(*cfg)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	errs := run(srv)

	select {
	case signals := <-sigs:
		srv.Logger.Info().Msgf("shutting down from signal: %v", signals)
		if err := srv.HttpServer.Shutdown(context.Background()); err != nil {
			srv.Logger.Err(err).Send()
		}
	case err := <-errs:
		srv.Logger.Error().Err(err).Msg("returning from ListenAndServe: %v")
	}
}

func run(srv http.Server) <-chan error {
	ch := make(chan error)
	go func() {
		srv.Logger.Info().Msgf("starting server on: %s", srv.HttpServer.Addr)
		ch <- srv.HttpServer.ListenAndServe()
		close(ch)
	}()
	return ch
}
