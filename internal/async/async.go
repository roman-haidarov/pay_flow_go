package async

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/google/uuid"
)

const (
	QueueDefault  = "default"
	TaskEmailSend = "email:send"
)

type ClientPayload struct {
	UserID uuid.UUID `json:"user_id"`
}

func NewClient(redisURL string) (*asynq.Client, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil { return nil, err }
	return asynq.NewClient(opt), nil
}

func NewServer(redisURL string, concurrency int) (*asynq.Server, *asynq.ServeMux, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil { return nil, nil, err }
	srv := asynq.NewServer(opt, asynq.Config{
		Concurrency:    concurrency,
		Queues:         map[string]int{QueueDefault: 1},
		StrictPriority: true,
	})
	mux := asynq.NewServeMux()
	return srv, mux, nil
}

func EnqueueUUid(cli *asynq.Client, p ClientPayload) (*asynq.TaskInfo, error) {
	b, _ := json.Marshal(p)
	t := asynq.NewTask(TaskEmailSend, b,
		asynq.Queue(QueueDefault),
		asynq.Timeout(30*time.Second),
		asynq.MaxRetry(4),
		asynq.Retention(24*time.Hour),
		asynq.Unique(1*time.Minute),
	)
	return cli.Enqueue(t)
}
