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

	if opt.ReadTimeout == 0 {
		opt.ReadTimeout = 1500 * time.Millisecond
	}
	if opt.WriteTimeout == 0 {
		opt.WriteTimeout = 1500 * time.Millisecond
	}
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 1500 * time.Millisecond
	}
	if opt.PoolSize == 0 {
		opt.PoolSize = 10
	}
	if opt.MinIdleConns == 0 {
		opt.MinIdleConns = 4
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

func (r *RedisCache) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	res, err := r.clt.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return res, err
}

func (r *RedisCache) Set(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	return r.clt.Set(ctx, key, value, 0).Err()
}

func (r *RedisCache) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	return r.clt.Del(ctx, key).Err()
}

func (r *RedisCache) Close() error { return r.clt.Close() }
