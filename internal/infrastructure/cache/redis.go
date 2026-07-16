package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/Otnielkalit/backend-pos/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

// NewRedis creates and validates a new Redis client.
// It pings Redis to confirm connectivity before returning.
func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("cache: invalid Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: ping failed: %w", err)
	}

	return client, nil
}

// MustNewRedis is like NewRedis but panics on error.
// Use only in main.go during startup.
func MustNewRedis(cfg config.RedisConfig) *redis.Client {
	client, err := NewRedis(cfg)
	if err != nil {
		panic(err)
	}
	return client
}
