package redis

import (
	"time"

	"github.com/slice-soft/ss-keel-core/contracts"
)

// PoolConfig controls the go-redis connection pool behaviour.
type PoolConfig struct {
	MaxActiveConns  int
	MinIdleConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

// Config is passed to New to configure the Redis client.
type Config struct {
	// URL is the Redis connection string.
	// Format: redis://[:password@]host[:port][/db-number]
	// Default: redis://localhost:6379
	URL      string
	SkipPing bool
	Pool     PoolConfig
	Logger   contracts.Logger
}

func (cfg *Config) withDefaults() {
	if cfg.URL == "" {
		cfg.URL = "redis://localhost:6379"
	}

	cfg.Pool.withDefaults()
}

func (p *PoolConfig) withDefaults() {
	if p.MaxActiveConns <= 0 {
		p.MaxActiveConns = 10
	}

	if p.MinIdleConns <= 0 {
		p.MinIdleConns = 2
	}

	if p.MaxIdleConns <= 0 {
		p.MaxIdleConns = 5
	}

	if p.ConnMaxIdleTime <= 0 {
		p.ConnMaxIdleTime = 5 * time.Minute
	}

	if p.ConnMaxLifetime <= 0 {
		p.ConnMaxLifetime = 30 * time.Minute
	}
}
