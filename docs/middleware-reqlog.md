# RequestAuditLog Middleware

`RequestAuditLog` emits a structured per-request log line that surfaces
fedproxy-specific context values — request-id and trace-id — making it
suitable for compliance and audit pipelines.

## Differences from AccessLog

| Feature | AccessLog | RequestAuditLog |
|---|---|---|
| Skippable paths | ✓ | ✓ |
| Request-ID | ✗ | ✓ |
| Trace-ID | ✗ | ✓ |
| Custom logger | ✗ | ✓ |

## Configuration

```go
type RequestLogConfig struct {
    Logger           *log.Logger // defaults to log.Default()
    SkipPaths        []string    // exact paths to suppress
    IncludeRequestID bool
    IncludeTraceID   bool
}
```

Use `DefaultRequestLogConfig()` to get a ready-to-use baseline.

## Usage

```go
cfg := middleware.DefaultRequestLogConfig()
cfg.SkipPaths = []string{"/healthz", "/_metrics"}

handler := middleware.RequestAuditLog(cfg)(proxyHandler)
```

## Log fields

Each line contains the following keys (printed via `log.Logger.Println`):

- `method` — HTTP verb
- `path` — request URI path
- `status` — response status code
- `bytes` — response body bytes written
- `duration_ms` — wall-clock time in milliseconds
- `remote` — client remote address
- `request_id` *(optional)* — value from `X-Request-ID` header or context
- `trace_id` *(optional)* — W3C traceparent trace-id from context

## Middleware order

Place `RequestAuditLog` **after** `RequestID` and `Tracing` so that both
identifiers are available in the context when the log line is emitted:

```
RequestID → Tracing → RequestAuditLog → … → proxy
```
