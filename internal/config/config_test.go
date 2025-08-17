package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoad_HappyPath(t *testing.T) {
	path := writeTemp(t, `
secret_key_base: "abc"
app_base_url: "http://localhost"
app_base_path: "/api"
port: 8080
redis_url: "redis://localhost:6379"
with_geocoder: true
location: "UTC"
log_level: 1
telemetry_endpoint: "http://collector:4318/v1/traces"
ipinfo_token: "tok"
consent_path: "/consent.yml"
otp:
  api_credential: "c"
  api_credential_test: "ct"
  api_salt: "s"
  api_url: "http://otp"
  api_url_test: "http://otpt"
  allowed_redirect_hosts: ["a.com","b.com"]
sender:
  api_url: "http://sms"
  api_user: "u"
  api_pass: "p"
  api_name: "n"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Port != 8080 || cfg.AppBasePath != "/api" || cfg.OTP.Credential != "c" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("no/such/file.yml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_BadYAML(t *testing.T) {
	path := writeTemp(t, "{bad: [yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}
