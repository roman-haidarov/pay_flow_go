package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type OTP struct {
	Credential     string   `env:"OTP_API_CREDENTIAL,required"`
	CredentialTest string   `env:"OTP_API_CREDENTIAL_TEST,required"`
	Salt           string   `env:"OTP_API_SALT,required"`
	URL            string   `env:"OTP_API_URL,required"`
	URLTest        string   `env:"OTP_API_URL_TEST,required"`
	AllowedHosts   []string `env:"OTP_ALLOWED_REDIRECT_HOSTS,required" envSeparator:","`
}

type Sender struct {
	ApiUrl string `env:"SENDER_API_URL,required"`
	User   string `env:"SENDER_API_USER,required"`
	Pass   string `env:"SENDER_API_PASS,required"`
	Name   string `env:"SENDER_API_NAME,required"`
}

type Config struct {
	Env string `env:"ENV,required"`

	// infra/runtime
	Port     			int    `env:"APP_PORT,required"`
	AppCpus  			string `env:"APP_CPUS,required"`
	RedisUrl 			string `env:"REDIS_URL,required"`
	RedisUrlLocal string `env:"REDIS_URL_LOCAL,required"`

	// app
	SecretKeyBase     string `env:"SECRET_KEY_BASE,required"`
	AppBaseURL        string `env:"APP_BASE_URL,required"`
	AppBasePath       string `env:"APP_BASE_PATH,required"`
	WithGeocoder      bool   `env:"WITH_GEOCODER,required"`
	Location          string `env:"LOCATION,required"`
	LogLevel          int    `env:"LOG_LEVEL,required"`
	TelemetryEndpoint string `env:"TELEMETRY_ENDPOINT,required"`
	IPInfoToken       string `env:"IPINFO_TOKEN,required"`
	ConsentPath       string `env:"CONSENT_PATH,required"`

	OTP    OTP
	Sender Sender
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	var c Config

	if err := env.ParseWithOptions(&c, env.Options{
			RequiredIfNoDef: true,
	}); err != nil {
			return nil, err
	}

	// sanity
	if c.Port < 1 || c.Port > 65535 {
		return nil, fmt.Errorf("invalid APP_PORT %d", c.Port)
	}
	if p := strings.TrimSpace(c.AppBasePath); p == "" || !strings.HasPrefix(p, "/") {
		return nil, fmt.Errorf("APP_BASE_PATH must start with '/'")
	}
	if _, err := time.LoadLocation(c.Location); err != nil {
		return nil, fmt.Errorf("invalid LOCATION %q: %w", c.Location, err)
	}

	c.RedisUrl = redisURL(&c)
	return &c, nil
}

func redisURL(cfg *Config) string {
	switch strings.ToLower(cfg.Env) {
	case "production":
		return cfg.RedisUrl
	default:
		return cfg.RedisUrlLocal
	}
}
