# Dedupe Middleware

The `Dedupe` middleware provides **idempotency-key-based request deduplication**.
When a client supplies the same key within the configured TTL window, the
original response is replayed from an in-memory cache without invoking the
upstream handler again.

This is useful for protecting non-idempotent backends (payments, order
submission, etc.) against duplicate submissions caused by retries.

## Configuration

```go
type DedupeConfig struct {
    TTL    time.Duration // how long to cache a response (default: 500ms)
    Header string        // request header carrying the key (default: X-Idempotency-Key)
}
```

Use `DefaultDedupeConfig()` for sensible defaults.

## Usage

```go
cfg := middleware.DefaultDedupeConfig()
cfg.TTL = 30 * time.Second

handler := middleware.Dedupe(cfg)(upstream)
```

## Behaviour

| Scenario | Result |
|---|---|
| No `X-Idempotency-Key` header | Request forwarded normally |
| Key seen, TTL not expired | Cached response replayed; `X-Dedupe-Replay: true` set |
| Key seen, TTL expired | Request forwarded; cache refreshed |
| Different keys | Each forwarded independently |

## Response Headers

- **`X-Dedupe-Replay: true`** — present on any response served from cache.

## Notes

- The cache is in-process only; it is not shared across multiple proxy
  instances. For distributed deduplication use an external store.
- Entries are not proactively evicted; memory usage is bounded by the number
  of unique keys seen within one TTL window.
- Compose `Dedupe` **inside** `RateLimit` so replayed responses do not consume
  rate-limit tokens on the inner handler.
