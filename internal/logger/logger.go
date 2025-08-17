package logger

import (
	"os"
	"pay_flow_go/internal/config"
	t "time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(level zerolog.Level, cfg *config.Config) {
	loc, err := t.LoadLocation(cfg.Location)
	if err != nil {
		loc = t.FixedZone(cfg.Location, 6*3600)
	}

	cw := zerolog.ConsoleWriter{
		Out:          os.Stdout,
		NoColor:      false,
		TimeFormat:   "2006-01-02 15:04:05",
		TimeLocation: loc,
	}

	zerolog.TimestampFunc = func() t.Time { return t.Now().In(loc) }
	log.Logger = zerolog.New(cw).With().Timestamp().Logger().With().Caller().Logger()

	zerolog.SetGlobalLevel(level)
}
