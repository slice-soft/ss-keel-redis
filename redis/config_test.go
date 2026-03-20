package redis

import (
	"testing"
	"time"
)

func TestConfigWithDefaults_AppliesExpectedDefaults(t *testing.T) {
	cfg := Config{}
	cfg.withDefaults()

	if cfg.URL != "redis://localhost:6379" {
		t.Fatalf("expected default URL redis://localhost:6379, got %q", cfg.URL)
	}
	if cfg.Pool.MaxActiveConns != 10 {
		t.Fatalf("expected MaxActiveConns 10, got %d", cfg.Pool.MaxActiveConns)
	}
	if cfg.Pool.MinIdleConns != 2 {
		t.Fatalf("expected MinIdleConns 2, got %d", cfg.Pool.MinIdleConns)
	}
	if cfg.Pool.MaxIdleConns != 5 {
		t.Fatalf("expected MaxIdleConns 5, got %d", cfg.Pool.MaxIdleConns)
	}
	if cfg.Pool.ConnMaxIdleTime != 5*time.Minute {
		t.Fatalf("expected ConnMaxIdleTime 5m, got %s", cfg.Pool.ConnMaxIdleTime)
	}
	if cfg.Pool.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("expected ConnMaxLifetime 30m, got %s", cfg.Pool.ConnMaxLifetime)
	}
}

func TestConfigWithDefaults_PreservesConfiguredValues(t *testing.T) {
	cfg := Config{
		URL: "redis://myhost:6380/1",
		Pool: PoolConfig{
			MaxActiveConns:  20,
			MinIdleConns:    4,
			MaxIdleConns:    8,
			ConnMaxIdleTime: 10 * time.Minute,
			ConnMaxLifetime: time.Hour,
		},
	}
	cfg.withDefaults()

	if cfg.URL != "redis://myhost:6380/1" {
		t.Fatalf("expected URL to be preserved, got %q", cfg.URL)
	}
	if cfg.Pool.MaxActiveConns != 20 {
		t.Fatalf("expected MaxActiveConns 20, got %d", cfg.Pool.MaxActiveConns)
	}
	if cfg.Pool.MinIdleConns != 4 {
		t.Fatalf("expected MinIdleConns 4, got %d", cfg.Pool.MinIdleConns)
	}
	if cfg.Pool.MaxIdleConns != 8 {
		t.Fatalf("expected MaxIdleConns 8, got %d", cfg.Pool.MaxIdleConns)
	}
	if cfg.Pool.ConnMaxIdleTime != 10*time.Minute {
		t.Fatalf("expected ConnMaxIdleTime 10m, got %s", cfg.Pool.ConnMaxIdleTime)
	}
	if cfg.Pool.ConnMaxLifetime != time.Hour {
		t.Fatalf("expected ConnMaxLifetime 1h, got %s", cfg.Pool.ConnMaxLifetime)
	}
}

func TestPoolConfigWithDefaults_NormalizesZeroValues(t *testing.T) {
	pool := PoolConfig{}
	pool.withDefaults()

	if pool.MaxActiveConns != 10 {
		t.Fatalf("expected MaxActiveConns 10, got %d", pool.MaxActiveConns)
	}
	if pool.MinIdleConns != 2 {
		t.Fatalf("expected MinIdleConns 2, got %d", pool.MinIdleConns)
	}
	if pool.MaxIdleConns != 5 {
		t.Fatalf("expected MaxIdleConns 5, got %d", pool.MaxIdleConns)
	}
	if pool.ConnMaxIdleTime != 5*time.Minute {
		t.Fatalf("expected ConnMaxIdleTime 5m, got %s", pool.ConnMaxIdleTime)
	}
	if pool.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("expected ConnMaxLifetime 30m, got %s", pool.ConnMaxLifetime)
	}
}
