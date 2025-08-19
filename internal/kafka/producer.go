package kafkaio

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"pay_flow_go/internal/config"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type SMS struct {
	UserID    uuid.UUID `json:"user_id"`
	Phone     string    `json:"phone"`
	IIN       string    `json:"iin"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Producer struct{ w *kafka.Writer }

func NewProducer(cfg *config.Kafka) *Producer {
	d := newDialer(cfg)
	w := newWriter(cfg, d)
	return &Producer{w: w}
}

func (p *Producer) Close() error { return p.w.Close() }

func (p *Producer) ProduceSMS(ctx context.Context, sms SMS) error {
	payload, err := json.Marshal(sms)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(sms.UserID.String()),
		Value: payload,
		Time:  time.Now().UTC(),
		Headers: []kafka.Header{
			{Key: "content-type", Value: []byte("application/json")},
			{Key: "event-type",   Value: []byte("sms")},
			{Key: "schema-ver",   Value: []byte("1")},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.w.WriteMessages(ctx, msg)
}

func newDialer(cfg *config.Kafka) *kafka.Dialer {
	d := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		ClientID:  cfg.Client.ClientID,
	}
	applyTLS(d, cfg.Security)
	applySASL(d, cfg.Security)
	return d
}

func newWriter(cfg *config.Kafka, d *kafka.Dialer) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(splitCSV(cfg.Client.BootstrapServers)...),
		Topic:                  cfg.Client.ProducerTopic,
		Balancer:               pickBalancer(cfg.Producer.Balancer),
		RequiredAcks:           pickAcks(cfg.Producer.RequiredAcks),
		BatchBytes:             batchBytes(cfg.Producer.BatchBytes),
		BatchTimeout:           time.Duration(cfg.Producer.LingerMs) * time.Millisecond,
		Compression:            pickCompression(cfg.Producer.Compression),
		MaxAttempts:            maxInt(cfg.Producer.Retries, 1),
		AllowAutoTopicCreation: cfg.Producer.AllowAutoTopicCreation,
		Transport: &kafka.Transport{
			TLS:  d.TLS,
			SASL: d.SASLMechanism,
		},
	}
}

func applyTLS(d *kafka.Dialer, sec config.KfkSecurity) {
	if sec.TLSEnable {
		d.TLS = &tls.Config{InsecureSkipVerify: sec.TLSInsecureSkipVerify}
	}
}
func applySASL(d *kafka.Dialer, sec config.KfkSecurity) {
	if !sec.SASLEnable {
		return
	}
	switch strings.ToUpper(strings.TrimSpace(sec.SASLMechanism)) {
	case "PLAIN":
		d.SASLMechanism = plain.Mechanism{Username: sec.SASLUsername, Password: sec.SASLPassword}
	case "SCRAM-SHA-512":
		m, _ := scram.Mechanism(scram.SHA512, sec.SASLUsername, sec.SASLPassword)
		d.SASLMechanism = m
	default: // SCRAM-SHA-256
		m, _ := scram.Mechanism(scram.SHA256, sec.SASLUsername, sec.SASLPassword)
		d.SASLMechanism = m
	}
}

func splitCSV(v []string) []string {
	out := make([]string, 0, len(v))
	for _, s := range v {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return []string{"broker:29092"}
	}
	return out
}

func pickAcks(s string) kafka.RequiredAcks {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "none":
		return kafka.RequireNone
	case "leader", "one":
		return kafka.RequireOne
	default:
		return kafka.RequireAll
	}
}

func pickCompression(s string) kafka.Compression {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "gzip":
		return kafka.Gzip
	case "lz4":
		return kafka.Lz4
	case "zstd":
		return kafka.Zstd
	default:
		return kafka.Snappy
	}
}

func pickBalancer(s string) kafka.Balancer {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "round-robin", "roundrobin":
		return &kafka.RoundRobin{}
	case "crc32":
		return &kafka.CRC32Balancer{}
	case "murmur2":
		return &kafka.Murmur2Balancer{}
	default:
		return &kafka.LeastBytes{}
	}
}

func batchBytes(v int64) int64 {
	if v > 0 {
		return v
	}
	return 1 << 20
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
