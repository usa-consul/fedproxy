# Throttle Middleware

The `Throttle` middleware limits the number of requests processed **concurrently**. Requests that arrive when all slots are occupied wait up to `QueueTimeout` for a slot to become available. If no slot is freed in time the middleware responds immediately with a configurable status code (default **503 Service Unavailable**).

This complements the `RateLimit` middleware (token-bucket, per-client) by enforcing a global cap on in-flight work, protecting downstream services from burst overload.

## Configuration

```go
type ThrottleConfig struct {
    MaxConcurrent int           // max simultaneous requests (default 100)
    QueueTimeout  time.Duration // how long to wait for a slot (default 5s)
    StatusCode    int           // status returned on timeout (default 503)
}
```

## Usage

```go
cfg := middleware.DefaultThrottleConfig()
cfg.MaxConcurrent = 50
cfg.QueueTimeout  = 2 * time.Second

handler := middleware.Throttle(cfg, myHandler)
```

## Behaviour

| Condition | Result |
|-----------|--------|
| Slot available | Request processed normally |
| No slot, within `QueueTimeout` | Request waits, then proceeds |
| No slot, timeout exceeded | `StatusCode` returned immediately |

## Chaining

Place `Throttle` **after** `RateLimit` and **before** the reverse proxy so that per-client rate limiting fires first, followed by the global concurrency cap:

```
RequestID → RateLimit → Throttle → ReverseProxy
```

## Notes

- The semaphore is released even if the upstream handler panics (use with `Recover`).
- `MaxConcurrent ≤ 0` falls back to the default of **100**.
- `QueueTimeout ≤ 0` falls back to the default of **5 s**.
