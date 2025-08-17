package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"pay_flow_go/internal/config"

	miniredis "github.com/alicebob/miniredis/v2"
)

func otlpSink() (*httptest.Server, string) {
	h := http.NewServeMux()

	h.HandleFunc("/v1/traces", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(h)
	return srv, srv.URL + "/v1/traces"
}

func TestServer_New_Run_Shutdown(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	otlp, endpoint := otlpSink()
	defer otlp.Close()

	cfg := &config.Config{
		RedisURL:          "redis://" + mr.Addr(),
		TelemetryEndpoint: endpoint,
		Location:          "UTC",
	}

	s, err := New(cfg)
	if err != nil {
		t.Fatalf("server.New error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.Run(ctx); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	s.Shutdown()
}
