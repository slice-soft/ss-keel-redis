package redis

import (
	"context"

	"github.com/slice-soft/ss-keel-core/contracts"
)

// HealthChecker implements contracts.HealthChecker for a Redis connection.
// Register it with app.RegisterHealthChecker(redis.NewHealthChecker(client))
// to expose the Redis status in GET /health.
type HealthChecker struct {
	client *Client
}

var _ contracts.HealthChecker = (*HealthChecker)(nil)

// NewHealthChecker returns a HealthChecker that pings Redis.
func NewHealthChecker(client *Client) *HealthChecker {
	return &HealthChecker{client: client}
}

// Name returns the key used in the /health response (e.g. "redis": "UP").
func (h *HealthChecker) Name() string {
	return "redis"
}

// Check pings Redis. Returns a non-nil error if the connection is unhealthy.
func (h *HealthChecker) Check(ctx context.Context) error {
	return h.client.rdb.Ping(ctx).Err()
}
