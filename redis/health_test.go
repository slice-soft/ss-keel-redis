package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func newHealthClient(t *testing.T) (*Client, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	client, err := New(Config{
		URL: "redis://" + mr.Addr(),
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	t.Cleanup(func() { _ = client.Close() })

	return client, mr
}

func TestHealthChecker_NameIsRedis(t *testing.T) {
	client, _ := newHealthClient(t)
	checker := NewHealthChecker(client)

	if checker.Name() != "redis" {
		t.Fatalf("expected name %q, got %q", "redis", checker.Name())
	}
}

func TestHealthChecker_CheckSucceedsWhenConnected(t *testing.T) {
	client, _ := newHealthClient(t)
	checker := NewHealthChecker(client)

	if err := checker.Check(context.Background()); err != nil {
		t.Fatalf("expected health check to succeed, got %v", err)
	}
}

func TestHealthChecker_CheckFailsWhenServerIsDown(t *testing.T) {
	client, mr := newHealthClient(t)
	checker := NewHealthChecker(client)

	mr.Close()

	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("expected health check to fail after server shutdown")
	}
}
