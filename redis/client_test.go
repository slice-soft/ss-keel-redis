package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// newTestClient starts an in-memory Redis server and returns a connected Client.
func newTestClient(t *testing.T) (*Client, *miniredis.Miniredis) {
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

func TestNew_ReturnsErrorForInvalidURL(t *testing.T) {
	_, err := New(Config{URL: "not-a-valid-url"})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestNew_SkipPingSkipsConnectivity(t *testing.T) {
	// Unreachable address — succeeds because ping is skipped.
	client, err := New(Config{
		URL:      "redis://127.0.0.1:1",
		SkipPing: true,
	})
	if err != nil {
		t.Fatalf("expected no error when SkipPing is true, got %v", err)
	}
	_ = client.Close()
}

func TestNew_LogsConnectionMessage(t *testing.T) {
	mr := miniredis.RunT(t)
	log := &testLogger{}

	client, err := New(Config{
		URL:    "redis://" + mr.Addr(),
		Logger: log,
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer client.Close()

	if len(log.infos) == 0 {
		t.Fatal("expected at least one info log message")
	}
}

func TestSetAndGet(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	want := []byte(`{"name":"Alice"}`)
	if err := client.Set(ctx, "user:1", want, time.Minute); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	got, err := client.Get(ctx, "user:1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestGet_ReturnsNilForMissingKey(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	val, err := client.Get(ctx, "does-not-exist")
	if err != nil {
		t.Fatalf("expected nil error for missing key, got %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil value for missing key, got %q", val)
	}
}

func TestSet_ZeroTTLMeansNoExpiration(t *testing.T) {
	client, mr := newTestClient(t)
	ctx := context.Background()

	if err := client.Set(ctx, "perm", []byte("value"), 0); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	// Advance miniredis clock — key should still be present.
	mr.FastForward(24 * time.Hour)

	val, err := client.Get(ctx, "perm")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if val == nil {
		t.Fatal("expected key to exist after 24h with zero TTL")
	}
}

func TestDelete(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	if err := client.Set(ctx, "to-delete", []byte("v"), time.Minute); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if err := client.Delete(ctx, "to-delete"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	val, err := client.Get(ctx, "to-delete")
	if err != nil {
		t.Fatalf("Get returned error after delete: %v", err)
	}
	if val != nil {
		t.Fatal("expected key to be absent after Delete")
	}
}

func TestExists(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()

	exists, err := client.Exists(ctx, "missing")
	if err != nil {
		t.Fatalf("Exists returned error: %v", err)
	}
	if exists {
		t.Fatal("expected Exists to return false for missing key")
	}

	if err := client.Set(ctx, "present", []byte("1"), time.Minute); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	exists, err = client.Exists(ctx, "present")
	if err != nil {
		t.Fatalf("Exists returned error: %v", err)
	}
	if !exists {
		t.Fatal("expected Exists to return true for present key")
	}
}

func TestSet_TTLExpiresKey(t *testing.T) {
	client, mr := newTestClient(t)
	ctx := context.Background()

	if err := client.Set(ctx, "expiry", []byte("v"), 5*time.Second); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	mr.FastForward(6 * time.Second)

	val, err := client.Get(ctx, "expiry")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if val != nil {
		t.Fatal("expected key to be expired")
	}
}

func TestRDB_ReturnsUnderlyingClient(t *testing.T) {
	client, _ := newTestClient(t)

	rdb := client.RDB()
	if rdb == nil {
		t.Fatal("expected RDB to return a non-nil go-redis client")
	}

	// Verify the underlying client works by calling a raw command.
	ctx := context.Background()
	if err := rdb.Set(ctx, "raw", []byte("ok"), 0).Err(); err != nil {
		t.Fatalf("raw Set via RDB returned error: %v", err)
	}
}

func TestClose(t *testing.T) {
	mr := miniredis.RunT(t)

	client, err := New(Config{URL: "redis://" + mr.Addr()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

// testLogger is a minimal contracts.Logger for test assertions.
type testLogger struct {
	infos []string
}

func (l *testLogger) Info(format string, args ...interface{}) {
	l.infos = append(l.infos, fmt.Sprintf(format, args...))
}

func (l *testLogger) Warn(string, ...interface{})  {}
func (l *testLogger) Error(string, ...interface{}) {}
func (l *testLogger) Debug(string, ...interface{}) {}
