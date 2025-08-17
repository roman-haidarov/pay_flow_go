package server

import (
	"context"
	"pay_flow_go/internal/cache"
	"pay_flow_go/internal/config"
	"github.com/rs/zerolog/log"
)

type Server struct {
	cfg *config.Config
	ch  cache.Store
	// api *api.API
}

func New(cfg *config.Config) (*Server, error) {
	rc, err := cache.New(cfg.RedisURL)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg: cfg,
		ch:  rc,
		// api: api.NewAPI(rc, reviews.NewService(rc)),
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	// Init Telemetry SDK.
	shutdown, err := setupOTelSDK(ctx, s.cfg.TelemetryEndpoint)
	if err != nil {
		return err
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Err(err).Msg("cannot shutdown OTel")
		}
	}()

	// return s.api.Serve(ctx)
	return nil
}

func (s *Server) Shutdown() {
	if s.ch != nil {
		if err := s.ch.Close(); err != nil {
			log.Err(err).Msg("redis close failed")
		}
	}
	log.Info().Msg("graceful server shutdown")
}
