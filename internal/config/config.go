package config

import (
	"strings"

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

type Kafka struct {
	Client   KfkClient
	Producer KfkProducer
	Consumer KfkConsumer
	Security KfkSecurity
}

type KfkClient struct {
	BootstrapServers    []string `env:"KAFKA_BOOTSTRAP_SERVERS,required"`
	ClientID            string   `env:"KAFKA_CLIENT_ID,required"`
	ProducerTopic       string   `env:"KAFKA_PRODUCER_TOPIC,required"`
	ConsumerTopic       string   `env:"KAFKA_CONSUMER_TOPIC,required"`
	ConsumerGroup       string   `env:"KAFKA_CONSUMER_GROUP,required"`
	ConsumerStartOffset string   `env:"KAFKA_CONSUMER_START_OFFSET,required"`
}
type KfkProducer struct {
	RequiredAcks           string `env:"KAFKA_PRODUCER_REQUIRED_ACKS,required"`
	LingerMs               int    `env:"KAFKA_PRODUCER_LINGER_MS,required"`
	BatchBytes             int64  `env:"KAFKA_PRODUCER_BATCH_BYTES,required"`
	Compression            string `env:"KAFKA_PRODUCER_COMPRESSION,required"`
	MaxInFlight            int    `env:"KAFKA_PRODUCER_MAX_IN_FLIGHT,required"`
	Retries                int    `env:"KAFKA_PRODUCER_RETRIES,required"`
	RetryBackoffMs         int    `env:"KAFKA_PRODUCER_RETRY_BACKOFF_MS,required"`
	Balancer               string `env:"KAFKA_PRODUCER_BALANCER,required"`
	AllowAutoTopicCreation bool   `env:"KAFKA_ALLOW_AUTO_TOPIC_CREATION,required"`
}

type KfkConsumer struct {
	MinBytes             int  	`env:"KAFKA_CONSUMER_MIN_BYTES,required"`
	MaxBytes             int  	`env:"KAFKA_CONSUMER_MAX_BYTES,required"`
	MaxWaitMs            int  	`env:"KAFKA_CONSUMER_MAX_WAIT_MS,required"`
	SessionTimeoutMs     int  	`env:"KAFKA_CONSUMER_SESSION_TIMEOUT_MS,required"`
	HeartbeatMs          int  	`env:"KAFKA_CONSUMER_HEARTBEAT_MS,required"`
	AutoCommit           bool 	`env:"KAFKA_CONSUMER_AUTO_COMMIT,required"`
	AutoCommitIntervalMs int  	`env:"KAFKA_CONSUMER_AUTO_COMMIT_INTERVAL_MS,required"`
	GroupID							 string `env:"KAFKA_CONSUMER_GROUP,required"`
	StartOffset					 string `env:"KAFKA_CONSUMER_OFFSET,required"`
	BatchSize						 int  	`env:"KAFKA_CONSUMER_BATCH,required"`
	TickMs 							 int  	`env:"KAFKA_CONSUMER_TICKS,required"`
}

type KfkSecurity struct {
	SASLEnable            bool   `env:"KAFKA_SASL_ENABLE,required"`
	SASLMechanism         string `env:"KAFKA_SASL_MECHANISM"            envDefault:"SCRAM-SHA-256"`
	SASLUsername          string `env:"KAFKA_SASL_USERNAME"`
	SASLPassword          string `env:"KAFKA_SASL_PASSWORD"`
	TLSEnable             bool   `env:"KAFKA_TLS_ENABLE"                 envDefault:"false"`
	TLSInsecureSkipVerify bool   `env:"KAFKA_TLS_INSECURE_SKIP_VERIFY"   envDefault:"false"`
}

type Config struct {
	Env string `env:"ENV,required"`

	Port          int    `env:"APP_PORT,required"`
	AppCpus       string `env:"APP_CPUS,required"`
	RedisUrl      string `env:"REDIS_URL,required"`
	RedisUrlLocal string `env:"REDIS_URL_LOCAL,required"`

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
	Kafka  Kafka
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	var c Config
	if err := env.ParseWithOptions(&c, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		return nil, err
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
