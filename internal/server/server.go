package server

import (
	"context"
	"pay_flow_go/internal/config"

	"github.com/rs/zerolog/log"
)

type Server struct {
	cfg     *config.Config
	// db      db.DB
	// api     *api.API
}

func New(cfg *config.Config) (*Server, error) {
	srv := &Server{
		cfg: cfg,
		// db:      db.New(cfg.DB),
		// api:     api.NewAPI(db.New(cfg.DB), reviews.NewService(db.New(cfg.DB))),
	}

	return srv, nil
}

func (s *Server) Run(ctx context.Context) error {
	// Init Telemetry SDK.
	shutdown, err := setupOTelSDK(ctx, s.cfg.TelemetryEndpoint)
	if err != nil {
		return err
	}
	defer func() {
		err = shutdown(ctx)
		if err != nil {
			log.Err(err).Msg("cannot shutdown OTel")
		}
	}()

	// return s.api.Serve(ctx)
	return nil
}

func (s *Server) Shutdown() {
	log.Info().Msg("graceful server shutdown")
}
