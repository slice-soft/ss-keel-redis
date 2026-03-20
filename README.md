<img src="https://cdn.slicesoft.dev/boat.svg" width="400" />

# ss-keel-redis
Keel is a Go framework for building REST APIs with modular
architecture, automatic OpenAPI, and built-in validation.

[![CI](https://github.com/slice-soft/ss-keel-redis/actions/workflows/ci.yml/badge.svg)](https://github.com/slice-soft/ss-keel-redis/actions)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/slice-soft/ss-keel-redis)](https://goreportcard.com/report/github.com/slice-soft/ss-keel-redis)
[![Go Reference](https://pkg.go.dev/badge/github.com/slice-soft/ss-keel-redis.svg)](https://pkg.go.dev/github.com/slice-soft/ss-keel-redis)
![License](https://img.shields.io/badge/License-MIT-green)
![Made in Colombia](https://img.shields.io/badge/Made%20in-Colombia-FCD116?labelColor=003893)


## Cache addon for Keel

`ss-keel-redis` adds Redis cache support to a [Keel](https://keel-go.dev) project via [go-redis v9](https://redis.uptrace.dev/).
It is the official addon for distributed caching in the Keel ecosystem and implements `contracts.Cache` from `ss-keel-core`.

---

## 🚀 Installation

```bash
keel add redis
```

The Keel CLI will:
1. Add `github.com/slice-soft/ss-keel-redis` as a dependency.
2. Create `cmd/setup_redis.go` and inject initialization code into `cmd/main.go`.
3. Add a `REDIS_URL` environment variable example to both `.env` and `.env.example`.

---

## ⚙️ Configuration

```go
client, err := ssredis.New(ssredis.Config{
    URL:    config.GetEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
    Logger: app.Logger(),
})
if err != nil {
    app.Logger().Error("failed to start redis: %v", err)
}
defer client.Close()
```

The `URL` field uses the standard Redis URL format: `redis://[:password@]host[:port][/db-number]`.

---

## 🔗 Connection pool

Pool defaults applied when not overridden:

| Parameter | Default |
|---|---|
| `MaxActiveConns` | 10 |
| `MinIdleConns` | 2 |
| `MaxIdleConns` | 5 |
| `ConnMaxIdleTime` | 5 min |
| `ConnMaxLifetime` | 30 min |

Override via `Config.Pool`:

```go
client, err := ssredis.New(ssredis.Config{
    URL:    config.GetEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
    Logger: app.Logger(),
    Pool: ssredis.PoolConfig{
        MaxActiveConns:  20,
        MinIdleConns:    5,
        ConnMaxLifetime: time.Hour,
    },
})
```

---

## 📦 Cache operations

`contracts.Cache` covers the four core operations:

```go
ctx := context.Background()

// Store a value with a TTL
err := client.Set(ctx, "user:123", []byte(`{"name":"Alice"}`), 5*time.Minute)

// Retrieve — returns nil, nil when the key does not exist
val, err := client.Get(ctx, "user:123")

// Remove a key
err = client.Delete(ctx, "user:123")

// Check existence without reading the value
exists, err := client.Exists(ctx, "user:123")
```

A zero TTL in `Set` means no expiration.

---

## 🔧 Advanced operations

Use `RDB()` to access the full go-redis client for pipelines, transactions, Lua scripts, and Pub/Sub:

```go
pipe := client.RDB().Pipeline()
pipe.Incr(ctx, "counter")
pipe.Expire(ctx, "counter", time.Hour)
_, err := pipe.Exec(ctx)
```

---

## ❤️ Health checker

Register the Redis connection in the Keel health endpoint:

```go
app.RegisterHealthChecker(ssredis.NewHealthChecker(client))
```

This exposes the Redis status under `GET /health`:

```json
{ "redis": "UP" }
```

---

## 🤚 CI/CD and releases

- **CI** runs on every pull request targeting `main` via `.github/workflows/ci.yml`.
- **Releases** are created automatically on merge to `main` via `.github/workflows/release.yml` using Release Please.

---

## 💡 Recommendations

* Use `REDIS_URL` for all environments; it keeps credentials out of code and plays well with secrets managers.
* Accept `contracts.Cache` in your services — not `*ssredis.Client` — so you can swap the implementation in tests.
* Register `NewHealthChecker` so Keel's `/health` endpoint always reflects real Redis connectivity.

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for setup and repository-specific rules.
The base workflow, commit conventions, and community standards live in [ss-community](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md).

## Community

| Document | |
|---|---|
| [CONTRIBUTING.md](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md) | Workflow, commit conventions, and PR guidelines |
| [GOVERNANCE.md](https://github.com/slice-soft/ss-community/blob/main/GOVERNANCE.md) | Decision-making, roles, and release process |
| [CODE_OF_CONDUCT.md](https://github.com/slice-soft/ss-community/blob/main/CODE_OF_CONDUCT.md) | Community standards |
| [VERSIONING.md](https://github.com/slice-soft/ss-community/blob/main/VERSIONING.md) | SemVer policy and breaking changes |
| [SECURITY.md](https://github.com/slice-soft/ss-community/blob/main/SECURITY.md) | How to report vulnerabilities |
| [MAINTAINERS.md](https://github.com/slice-soft/ss-community/blob/main/MAINTAINERS.md) | Active maintainers |

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- Website: [keel-go.dev](https://keel-go.dev)
- GitHub: [github.com/slice-soft/ss-keel-cli](https://github.com/slice-soft/ss-keel-cli)
- Documentation: [docs.keel-go.dev](https://docs.keel-go.dev)

---

Made by [SliceSoft](https://slicesoft.dev) — Colombia 💙
