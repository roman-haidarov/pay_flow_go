package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"pay_flow_go/internal/config"
	"pay_flow_go/internal/logger"
	"pay_flow_go/internal/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "go.uber.org/automaxprocs"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	if err != nil {
		panic("can not load config")
	}

	lvl := zerolog.Level(cfg.LogLevel)
	logger.Init(lvl, cfg)
	log.Info().Msgf("config loaded: %+v", cfg)
	srv, err := server.New(cfg)

	if err != nil {
		log.Fatal().Err(err).Msg("can not create server")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := srv.Run(ctx); err != nil {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	log.Info().Msg("Server started")
	<-ctx.Done()
	srv.Shutdown()
}
