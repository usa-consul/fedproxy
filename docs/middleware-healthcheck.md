# Health Check Middleware

The `HealthCheck` middleware intercepts requests to a configurable path (default `/healthz`) and returns a JSON health payload. All other requests pass through to the next handler unchanged.

## Configuration

```go
type HealthCheckConfig struct {
    Path        string        // URL path (default: /healthz)
    UpstreamURL string        // Optional upstream to probe
    Timeout     time.Duration // Probe timeout (default: 3s)
}
```

Use `DefaultHealthCheckConfig()` for sensible defaults.

## Response Format

```json
{
  "status": "ok",
  "upstream": "ok",
  "timestamp": "2024-01-15T12:00:00Z"
}
```

| Field      | Values                  | Notes                          |
|------------|-------------------------|--------------------------------|
| `status`   | `ok` \| `degraded`      | Reflects upstream probe result |
| `upstream` | `ok` \| `unreachable`   | Omitted when no URL configured |
| `timestamp`| RFC 3339 UTC string     | Time of the health check       |

## HTTP Status Codes

- **200 OK** — proxy is healthy; upstream (if configured) responded successfully
- **503 Service Unavailable** — upstream probe failed or returned 5xx

## Usage

```go
cfg := middleware.DefaultHealthCheckConfig()
cfg.UpstreamURL = "http://backend:8080/health"

handler := middleware.HealthCheck(cfg, proxyHandler)
```

## Notes

- The endpoint always sets `Cache-Control: no-store` to prevent stale health data.
- The upstream probe uses a short-lived `http.Client` scoped to the request; it does not reuse the proxy's transport.
- Exempt the health path from auth middleware to allow unauthenticated liveness checks from load balancers.
