package logger

import (
	"testing"

	"pay_flow_go/internal/config"

	"github.com/rs/zerolog"
)

func TestInit_SetsGlobalLevel(t *testing.T) {
	cfg := &config.Config{Location: "UTC"}
	Init(zerolog.InfoLevel, cfg)
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Fatalf("global level: want INFO, got %v", zerolog.GlobalLevel())
	}
}
