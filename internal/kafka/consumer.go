// internal/kafka/consumer.go
package kafkaio

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"pay_flow_go/internal/config"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

// BatchItem — элемент батча. Поле commit используется как "токен"
// для подтверждения (partition+offset); наружу его не выносим.
type BatchItem struct {
	SMS    SMS
	commit kafka.Message
}

type Consumer struct {
	r         *kafka.Reader
	tp        string
	batchSize int
	tick      time.Duration
	fetchWait time.Duration
}

func NewConsumer(cfg *config.Kafka) *Consumer {
	d := newDialer(cfg)
	topic := strings.TrimSpace(cfg.Client.ConsumerTopic)

	rc := kafka.ReaderConfig{
		Brokers:           splitCSV(cfg.Client.BootstrapServers),
		GroupID:           cfg.Consumer.GroupID,
		Topic:             topic,
		Dialer:            d,
		MinBytes:          cfg.Consumer.MinBytes,
		MaxBytes:          cfg.Consumer.MaxBytes,
		MaxWait:           time.Duration(cfg.Consumer.MaxWaitMs) * time.Millisecond,
		SessionTimeout:    time.Duration(cfg.Consumer.SessionTimeoutMs) * time.Millisecond,
		HeartbeatInterval: time.Duration(cfg.Consumer.HeartbeatMs) * time.Millisecond,
		CommitInterval: func() time.Duration {
			if cfg.Consumer.AutoCommit {
				return time.Duration(cfg.Consumer.AutoCommitIntervalMs) * time.Millisecond
			}
			return 0 // только ручные коммиты
		}(),
	}
	rc.StartOffset = parseStartOffset(cfg.Consumer.StartOffset)

	return &Consumer{
		r:         kafka.NewReader(rc),
		tp:        topic,
		batchSize: cfg.Consumer.BatchSize,
		tick:      time.Duration(cfg.Consumer.TickMs) * time.Millisecond,
		fetchWait: time.Duration(cfg.Consumer.MaxWaitMs) * time.Millisecond, // ожидание до 1-го сообщения в тик
	}
}

// Start — неблокирующий запуск тикового чтения.
// Раз в tick собирает батч и вызывает handler.
// handler должен вернуть индексы успешно обработанных элементов (okIdx).
func (c *Consumer) Start(ctx context.Context, handler func(context.Context, []BatchItem) ([]int, error)) {
	go c.run(ctx, handler)
}

func (c *Consumer) run(ctx context.Context, handler func(context.Context, []BatchItem) ([]int, error)) {
	ticker := time.NewTicker(c.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			items, err := c.PollBatch(ctx, c.fetchWait)
			if err != nil {
				log.Error().Err(err).Msg("poll batch failed")
				continue
			}
			if len(items) == 0 {
				continue
			}

			okIdx, err := handler(ctx, items)
			if err != nil {
				log.Error().Err(err).Msg("handler error")
			}
			if err := c.CommitContiguous(ctx, items, okIdx); err != nil {
				log.Warn().Err(err).Msg("commit contiguous failed")
			}
		}
	}
}

// PollBatch делает один "тик": пытается набрать до batchSize сообщений.
// Для первого сообщения ждёт до fetchWait, затем больше не ждёт.
func (c *Consumer) PollBatch(ctx context.Context, fetchWait time.Duration) ([]BatchItem, error) {
	maxN := c.batchSize
	items := make([]BatchItem, 0, maxN)

	firstDeadline, cancelFirst := context.WithTimeout(ctx, fetchWait)
	defer cancelFirst()

	waitCtx := firstDeadline
	for len(items) < maxN {
		m, err := c.r.FetchMessage(waitCtx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				break
			}
			log.Warn().Err(err).Msg("kafka fetch failed; continue")
			time.Sleep(50 * time.Millisecond)
			break
		}

		// после первого успешного чтения — больше не ждём
		waitCtx = ctx

		// опциональный фильтр по заголовку
		if !isSMSMessage(m.Headers) {
			if err := c.r.CommitMessages(ctx, m); err != nil {
				log.Warn().Err(err).Msg("commit skipped non-sms failed")
			}
			continue
		}

		var sms SMS
		if err := json.Unmarshal(m.Value, &sms); err != nil {
			// битное сообщение: логируем и сдвигаем оффсет (в бою — лучше DLQ)
			log.Error().Err(err).Msg("decode failed; committing to skip")
			_ = c.r.CommitMessages(ctx, m)
			continue
		}

		items = append(items, BatchItem{SMS: sms, commit: m})
	}

	return items, nil
}

// Commit — явный коммит набора элементов (если уверен, что все успешны).
func (c *Consumer) Commit(ctx context.Context, batch []BatchItem) error {
	if len(batch) == 0 {
		return nil
	}
	msgs := make([]kafka.Message, 0, len(batch))
	for _, it := range batch {
		msgs = append(msgs, it.commit)
	}
	return c.r.CommitMessages(ctx, msgs...)
}

// CommitOne — коммит одного элемента.
func (c *Consumer) CommitOne(ctx context.Context, it BatchItem) error {
	return c.r.CommitMessages(ctx, it.commit)
}

// CommitContiguous — коммитит максимум непрерывный префикс успешных
// сообщений в КАЖДОЙ партиции (до первого неуспеха). Это сохраняет FIFO.
func (c *Consumer) CommitContiguous(ctx context.Context, batch []BatchItem, okIdx []int) error {
	if len(batch) == 0 || len(okIdx) == 0 {
		return nil
	}
	ok := make(map[int]struct{}, len(okIdx))
	for _, i := range okIdx {
		ok[i] = struct{}{}
	}

	type partState struct {
		blocked  bool
		toCommit *kafka.Message // последний подряд успешный
	}
	parts := map[int]*partState{}

	for i, it := range batch {
		p := it.commit.Partition
		st := parts[p]
		if st == nil {
			st = &partState{}
			parts[p] = st
		}
		if st.blocked {
			continue
		}
		if _, success := ok[i]; success {
			m := it.commit
			st.toCommit = &m
		} else {
			st.blocked = true
		}
	}

	msgs := make([]kafka.Message, 0, len(parts))
	for _, st := range parts {
		if st.toCommit != nil {
			msgs = append(msgs, *st.toCommit)
		}
	}
	if len(msgs) == 0 {
		return nil
	}
	return c.r.CommitMessages(ctx, msgs...)
}

func (c *Consumer) Close() error { return c.r.Close() }

// --- helpers ---

func parseStartOffset(s string) int64 {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "earliest", "first":
		return kafka.FirstOffset
	case "latest", "last":
		return kafka.LastOffset
	default:
		return kafka.LastOffset
	}
}

func isSMSMessage(h []kafka.Header) bool {
	for _, hd := range h {
		if strings.EqualFold(hd.Key, "event-type") && strings.EqualFold(string(hd.Value), "sms") {
			return true
		}
	}
	// по умолчанию считаем sms-событием
	return true
}
