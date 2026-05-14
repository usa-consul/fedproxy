# Metrics Middleware

The `Metrics` middleware collects lightweight per-request counters and exposes
them as a JSON endpoint, making it easy to integrate with monitoring dashboards
or health-check scripts without an external dependency.

## Configuration

```go
cfg := middleware.DefaultMetricsConfig()
// cfg.Path == "/_metrics"
```

| Field  | Type     | Default      | Description                              |
|--------|----------|--------------|------------------------------------------|
| `Path` | `string` | `/_metrics`  | URL path that serves the JSON snapshot.  |

## Usage

```go
mux := http.NewServeMux()
mux.Handle("/", proxyHandler)

handler := middleware.Metrics(middleware.DefaultMetricsConfig())(mux)
http.ListenAndServe(":8080", handler)
```

## JSON Response

```json
{
  "total_requests": 1042,
  "total_errors": 3,
  "total_bytes_sent": 2097152,
  "uptime": "2h15m0s",
  "avg_latency_ms": 4.72
}
```

| Field             | Description                                      |
|-------------------|--------------------------------------------------|
| `total_requests`  | All proxied requests (excludes metrics endpoint) |
| `total_errors`    | Responses with HTTP status >= 500                |
| `total_bytes_sent`| Sum of response body bytes                       |
| `uptime`          | Time since the process started                   |
| `avg_latency_ms`  | Mean handler latency in milliseconds             |

## Notes

- Counters are **atomic** and safe for concurrent use.
- The metrics endpoint itself is **not** counted as a request.
- The response always includes `Cache-Control: no-store`.
- Counters reset when the process restarts (in-memory only).
