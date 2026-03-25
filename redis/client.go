package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/slice-soft/ss-keel-core/contracts"
)

// Client wraps go-redis and implements contracts.Cache.
type Client struct {
	rdb    *redis.Client
	logger contracts.Logger
	events chan contracts.PanelEvent
}

var _ contracts.Cache = (*Client)(nil)

// New creates a new Redis Client and optionally pings the server.
func New(cfg Config) (*Client, error) {
	cfg.withDefaults()

	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("ss-keel-redis: invalid REDIS_URL: %w", err)
	}

	opts.PoolSize        = cfg.Pool.MaxActiveConns
	opts.MinIdleConns    = cfg.Pool.MinIdleConns
	opts.MaxIdleConns    = cfg.Pool.MaxIdleConns
	opts.ConnMaxIdleTime = cfg.Pool.ConnMaxIdleTime
	opts.ConnMaxLifetime = cfg.Pool.ConnMaxLifetime

	rdb := redis.NewClient(opts)

	if !cfg.SkipPing {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			_ = rdb.Close()
			return nil, fmt.Errorf("ss-keel-redis: ping failed: %w", err)
		}
	}

	if cfg.Logger != nil {
		cfg.Logger.Info("redis connected [url=%s]", cfg.URL)
	}

	return &Client{rdb: rdb, logger: cfg.Logger, events: make(chan contracts.PanelEvent, 256)}, nil
}

// tryEmit sends an event to the panel channel without blocking.
func (c *Client) tryEmit(e contracts.PanelEvent) {
	select {
	case c.events <- e:
	default:
	}
}

// Get retrieves the raw bytes stored at key.
// Returns nil, nil when the key does not exist.
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	start := time.Now()
	val, err := c.rdb.Get(ctx, key).Bytes()
	hit := true
	if err == redis.Nil {
		hit = false
		err = nil
		val = nil
	}
	c.tryEmit(contracts.PanelEvent{
		Timestamp: time.Now(),
		AddonID:   "redis",
		Label:     "get",
		Level:     "info",
		Detail: map[string]any{
			"hit":         hit,
			"duration_ms": time.Since(start).Milliseconds(),
		},
	})
	return val, err
}

// Set stores value at key with the given TTL. A zero TTL means no expiration.
func (c *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	start := time.Now()
	err := c.rdb.Set(ctx, key, value, ttl).Err()
	c.tryEmit(contracts.PanelEvent{
		Timestamp: time.Now(),
		AddonID:   "redis",
		Label:     "set",
		Level:     "info",
		Detail: map[string]any{
			"duration_ms": time.Since(start).Milliseconds(),
			"ttl_ms":      ttl.Milliseconds(),
		},
	})
	return err
}

// Delete removes key from the cache.
func (c *Client) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := c.rdb.Del(ctx, key).Err()
	c.tryEmit(contracts.PanelEvent{
		Timestamp: time.Now(),
		AddonID:   "redis",
		Label:     "delete",
		Level:     "info",
		Detail: map[string]any{
			"duration_ms": time.Since(start).Milliseconds(),
		},
	})
	return err
}

// Exists reports whether key is present in the cache.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	n, err := c.rdb.Exists(ctx, key).Result()
	found := n > 0
	c.tryEmit(contracts.PanelEvent{
		Timestamp: time.Now(),
		AddonID:   "redis",
		Label:     "exists",
		Level:     "info",
		Detail: map[string]any{
			"duration_ms": time.Since(start).Milliseconds(),
			"found":       found,
		},
	})
	return found, err
}

// RDB returns the underlying go-redis client for advanced operations.
func (c *Client) RDB() *redis.Client {
	return c.rdb
}

// Close closes the connection pool.
func (c *Client) Close() error {
	return c.rdb.Close()
}
