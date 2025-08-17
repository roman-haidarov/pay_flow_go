package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type OTP struct {
	Credential     string   `yaml:"api_credential"`
	CredentialTest string   `yaml:"api_credential_test"`
	Salt           string   `yaml:"api_salt"`
	URL            string   `yaml:"api_url"`
	URLTest        string   `yaml:"api_url_test"`
	AllowedHosts   []string `yaml:"allowed_redirect_hosts"`
}

type Sender struct {
	APIURL string `yaml:"api_url"`
	User   string `yaml:"api_user"`
	Pass   string `yaml:"api_pass"`
	Name   string `yaml:"api_name"`
}

type Config struct {
	SecretKeyBase     string `yaml:"secret_key_base"`
	AppBaseURL        string `yaml:"app_base_url"`
	AppBasePath       string `yaml:"app_base_path"`
	Port              int    `yaml:"port"`
	RedisURL          string `yaml:"redis_url"`
	WithGeocoder      bool   `yaml:"with_geocoder"`
	Location          string `yaml:"location"`
	LogLevel          int    `yaml:"log_level"`
	TelemetryEndpoint string `yaml:"telemetry_endpoint"`
	IPInfoToken       string `yaml:"ipinfo_token"`
	ConsentPath       string `yaml:"consent_path"`
	OTP               OTP    `yaml:"otp"`
	Sender            Sender `yaml:"sender"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("PAY_FLOW_GO_CONFIG")
	}
	var c Config
	b, err := os.ReadFile(path)

	if err != nil {
		return &c, err
	}

	if err := yaml.Unmarshal(b, &c); err != nil {
		return &c, err
	}

	return &c, nil
}
