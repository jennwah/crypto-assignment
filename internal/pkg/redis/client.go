package redis

import (
	"context"
	"fmt"

	"github.com/jennwah/crypto-assignment/internal/config"
	"github.com/redis/go-redis/v9"
)

// NewClient returns Redis Client connection.
func NewClient(cfg config.Redis) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPass,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}
