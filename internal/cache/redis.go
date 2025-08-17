package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNotImplemented = errors.New("not implemented")

type RedisCache struct {
	clt *redis.Client
}

var _ Store = (*RedisCache)(nil)

func New(dsn string) (*RedisCache, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("redis dsn is empty")
	}

	var (
		opt *redis.Options
		err error
	)

	if strings.Contains(dsn, "://") {
		opt, err = redis.ParseURL(dsn)
		if err != nil {
			return nil, err
		}
	} else {
		opt = &redis.Options{Addr: dsn}
	}

	clt := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := clt.Ping(ctx).Err(); err != nil {
		_ = clt.Close()
		return nil, err
	}

	return &RedisCache{clt: clt}, nil
}

func (r *RedisCache) Get(key string) (string, error)  { return "", ErrNotImplemented }
func (r *RedisCache) Set(key, value string) error     { return ErrNotImplemented }
func (r *RedisCache) Delete(key string) error         { return ErrNotImplemented }
func (r *RedisCache) Close() error                    { return r.clt.Close() }
