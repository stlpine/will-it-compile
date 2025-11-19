package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/stlpine/will-it-compile/internal/config"
)

// Client wraps the Redis client with our configuration.
type Client struct {
	client *redis.Client
	ctx    context.Context
}

// NewClient creates a new Redis client with the given configuration.
func NewClient(cfg config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MaxRetries:   cfg.MaxRetries,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", cfg.Addr, err)
	}

	return &Client{
		client: client,
		ctx:    ctx,
	}, nil
}

// GetClient returns the underlying Redis client.
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// Ping tests the connection to Redis.
func (c *Client) Ping() error {
	return c.client.Ping(c.ctx).Err()
}
