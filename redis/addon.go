package redis

import "github.com/slice-soft/ss-keel-core/contracts"

// Compile-time assertions.
var (
	_ contracts.Addon        = (*Client)(nil)
	_ contracts.Debuggable   = (*Client)(nil)
	_ contracts.Manifestable = (*Client)(nil)
)

// ID implements contracts.Addon.
func (c *Client) ID() string { return "redis" }

// PanelID implements contracts.Debuggable.
func (c *Client) PanelID() string { return "redis" }

// PanelLabel implements contracts.Debuggable.
func (c *Client) PanelLabel() string { return "Cache (Redis)" }

// PanelEvents implements contracts.Debuggable.
func (c *Client) PanelEvents() <-chan contracts.PanelEvent { return c.events }

// Manifest implements contracts.Manifestable.
func (c *Client) Manifest() contracts.AddonManifest {
	return contracts.AddonManifest{
		ID:           "redis",
		Version:      "1.0.0",
		Capabilities: []string{"cache"},
		Resources:    []string{"redis"},
		EnvVars: []contracts.EnvVar{
			{
				Key:         "REDIS_URL",
				ConfigKey:   "redis.url",
				Description: "Redis connection URL",
				Required:    false,
				Secret:      false,
				Default:     "redis://localhost:6379",
				Source:      "redis",
			},
		},
	}
}

// RegisterWithPanel registers the client as a debuggable addon with the panel registry.
func (c *Client) RegisterWithPanel(r contracts.PanelRegistry) {
	r.RegisterAddon(c)
}
