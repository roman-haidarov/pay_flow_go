package main

import (
	"context"
	"os"
	"os/signal"
	"pay_flow_go/internal/config"
	"pay_flow_go/internal/logger"
	"pay_flow_go/internal/server"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load("config.yml")
	if err != nil {
		panic("can not load config")
	}
	logger.Init(zerolog.DebugLevel, cfg)
	log.Info().Msgf("config loaded: %+v", cfg)
	srv, err := server.New(cfg)

	if err != nil {
		log.Fatal().Err(err).Msg("can not create server")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
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
