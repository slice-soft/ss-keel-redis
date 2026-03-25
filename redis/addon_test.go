package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/slice-soft/ss-keel-core/contracts"
)

// Compile-time interface assertions (redundant with addon.go but explicit in tests).
var (
	_ contracts.Addon        = (*Client)(nil)
	_ contracts.Debuggable   = (*Client)(nil)
	_ contracts.Manifestable = (*Client)(nil)
)

// newTestClientForAddon creates a Client backed by miniredis for addon tests.
func newTestClientForAddon(t *testing.T) *Client {
	t.Helper()
	mr := miniredis.RunT(t)
	client, err := New(Config{URL: "redis://" + mr.Addr()})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// drainEvent reads the next event from the channel with a short timeout.
func drainEvent(t *testing.T, ch <-chan contracts.PanelEvent) contracts.PanelEvent {
	t.Helper()
	select {
	case e := <-ch:
		return e
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for panel event")
		return contracts.PanelEvent{}
	}
}

// --- Identity and label tests ---

func TestAddon_ID(t *testing.T) {
	c := newTestClientForAddon(t)
	if got := c.ID(); got != "redis" {
		t.Fatalf("expected ID %q, got %q", "redis", got)
	}
}

func TestAddon_PanelID(t *testing.T) {
	c := newTestClientForAddon(t)
	if got := c.PanelID(); got != "redis" {
		t.Fatalf("expected PanelID %q, got %q", "redis", got)
	}
}

func TestAddon_PanelLabel(t *testing.T) {
	c := newTestClientForAddon(t)
	if got := c.PanelLabel(); got != "Cache (Redis)" {
		t.Fatalf("expected PanelLabel %q, got %q", "Cache (Redis)", got)
	}
}

// --- Manifest tests ---

func TestAddon_Manifest_ID(t *testing.T) {
	c := newTestClientForAddon(t)
	m := c.Manifest()
	if m.ID != "redis" {
		t.Fatalf("expected manifest ID %q, got %q", "redis", m.ID)
	}
}

func TestAddon_Manifest_Capabilities(t *testing.T) {
	c := newTestClientForAddon(t)
	m := c.Manifest()
	if len(m.Capabilities) != 1 || m.Capabilities[0] != "cache" {
		t.Fatalf("expected capabilities [cache], got %v", m.Capabilities)
	}
}

func TestAddon_Manifest_Resources(t *testing.T) {
	c := newTestClientForAddon(t)
	m := c.Manifest()
	if len(m.Resources) != 1 || m.Resources[0] != "redis" {
		t.Fatalf("expected resources [redis], got %v", m.Resources)
	}
}

func TestAddon_Manifest_EnvVars(t *testing.T) {
	c := newTestClientForAddon(t)
	m := c.Manifest()
	if len(m.EnvVars) == 0 {
		t.Fatal("expected at least one EnvVar in manifest")
	}
	ev := m.EnvVars[0]
	if ev.Key != "REDIS_URL" {
		t.Fatalf("expected EnvVar key REDIS_URL, got %q", ev.Key)
	}
	if ev.ConfigKey != "redis.url" {
		t.Fatalf("expected ConfigKey redis.url, got %q", ev.ConfigKey)
	}
	if ev.Required {
		t.Fatal("expected REDIS_URL to be optional")
	}
	if ev.Secret {
		t.Fatal("expected REDIS_URL to not be secret")
	}
	if ev.Default != "redis://localhost:6379" {
		t.Fatalf("expected REDIS_URL default redis://localhost:6379, got %q", ev.Default)
	}
}

// --- PanelEvents channel test ---

func TestAddon_PanelEvents_ReturnsChannel(t *testing.T) {
	c := newTestClientForAddon(t)
	ch := c.PanelEvents()
	if ch == nil {
		t.Fatal("expected PanelEvents to return a non-nil channel")
	}
}

// --- tryEmit tests ---

func TestTryEmit_SendsEventToChannel(t *testing.T) {
	c := &Client{events: make(chan contracts.PanelEvent, 256)}
	e := contracts.PanelEvent{
		Timestamp: time.Now(),
		AddonID:   "redis",
		Label:     "test",
		Level:     "info",
		Detail:    map[string]any{"key": "value"},
	}
	c.tryEmit(e)

	select {
	case got := <-c.events:
		if got.Label != "test" {
			t.Fatalf("expected label %q, got %q", "test", got.Label)
		}
	default:
		t.Fatal("expected an event in the channel after tryEmit")
	}
}

func TestTryEmit_DoesNotBlockWhenChannelFull(t *testing.T) {
	c := &Client{events: make(chan contracts.PanelEvent, 1)}

	// Fill the channel.
	c.tryEmit(contracts.PanelEvent{Label: "first"})

	done := make(chan struct{})
	go func() {
		// This must not block.
		c.tryEmit(contracts.PanelEvent{Label: "overflow"})
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(200 * time.Millisecond):
		t.Fatal("tryEmit blocked on a full channel")
	}
}

// --- Instrumented method event tests (using miniredis) ---

func TestGet_Miss_EmitsEventWithHitFalse(t *testing.T) {
	c := newTestClientForAddon(t)
	ctx := context.Background()

	_, _ = c.Get(ctx, "missing-key")

	e := drainEvent(t, c.PanelEvents())
	if e.Label != "get" {
		t.Fatalf("expected label %q, got %q", "get", e.Label)
	}
	if e.AddonID != "redis" {
		t.Fatalf("expected addonID %q, got %q", "redis", e.AddonID)
	}
	hit, ok := e.Detail["hit"].(bool)
	if !ok {
		t.Fatalf("expected detail[hit] to be bool, got %T", e.Detail["hit"])
	}
	if hit {
		t.Fatal("expected hit=false for a cache miss")
	}
}

func TestGet_Hit_EmitsEventWithHitTrue(t *testing.T) {
	c := newTestClientForAddon(t)
	ctx := context.Background()

	if err := c.Set(ctx, "present", []byte("v"), time.Minute); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	// Drain the Set event.
	drainEvent(t, c.PanelEvents())

	_, _ = c.Get(ctx, "present")

	e := drainEvent(t, c.PanelEvents())
	if e.Label != "get" {
		t.Fatalf("expected label %q, got %q", "get", e.Label)
	}
	hit, ok := e.Detail["hit"].(bool)
	if !ok {
		t.Fatalf("expected detail[hit] to be bool, got %T", e.Detail["hit"])
	}
	if !hit {
		t.Fatal("expected hit=true for a cache hit")
	}
}

func TestSet_EmitsEvent(t *testing.T) {
	c := newTestClientForAddon(t)
	ctx := context.Background()

	if err := c.Set(ctx, "k", []byte("v"), 10*time.Second); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	e := drainEvent(t, c.PanelEvents())
	if e.Label != "set" {
		t.Fatalf("expected label %q, got %q", "set", e.Label)
	}
	if e.AddonID != "redis" {
		t.Fatalf("expected addonID %q, got %q", "redis", e.AddonID)
	}
	if _, ok := e.Detail["duration_ms"]; !ok {
		t.Fatal("expected detail to contain duration_ms")
	}
	if _, ok := e.Detail["ttl_ms"]; !ok {
		t.Fatal("expected detail to contain ttl_ms")
	}
}

func TestDelete_EmitsEvent(t *testing.T) {
	c := newTestClientForAddon(t)
	ctx := context.Background()

	// Set a key first; drain that event.
	_ = c.Set(ctx, "del-key", []byte("v"), time.Minute)
	drainEvent(t, c.PanelEvents())

	if err := c.Delete(ctx, "del-key"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	e := drainEvent(t, c.PanelEvents())
	if e.Label != "delete" {
		t.Fatalf("expected label %q, got %q", "delete", e.Label)
	}
	if _, ok := e.Detail["duration_ms"]; !ok {
		t.Fatal("expected detail to contain duration_ms")
	}
}

// --- Verify no key is emitted in events ---

func TestEvents_DoNotContainKey(t *testing.T) {
	c := newTestClientForAddon(t)
	ctx := context.Background()

	sensitiveKey := "user:secret-123"
	_ = c.Set(ctx, sensitiveKey, []byte("data"), time.Minute)
	e := drainEvent(t, c.PanelEvents())

	for k, v := range e.Detail {
		if k == "key" {
			t.Fatalf("event detail must not contain 'key' field, found value: %v", v)
		}
		if s, ok := v.(string); ok && s == sensitiveKey {
			t.Fatalf("event detail must not contain the key value in field %q", k)
		}
	}
}

// --- RegisterWithPanel test ---

func TestRegisterWithPanel(t *testing.T) {
	c := newTestClientForAddon(t)
	reg := &mockPanelRegistry{}
	c.RegisterWithPanel(reg)
	if reg.registered == nil {
		t.Fatal("expected RegisterAddon to be called")
	}
	if reg.registered.PanelID() != "redis" {
		t.Fatalf("expected registered panelID %q, got %q", "redis", reg.registered.PanelID())
	}
}

type mockPanelRegistry struct {
	registered contracts.Debuggable
}

func (m *mockPanelRegistry) RegisterAddon(d contracts.Debuggable) {
	m.registered = d
}

// Silence unused import if go-redis is not directly used elsewhere in this file.
var _ = redis.Nil
