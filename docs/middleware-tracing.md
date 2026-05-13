# Tracing Middleware

The `Tracing` middleware provides lightweight distributed tracing support using
the [W3C Traceparent](https://www.w3.org/TR/trace-context/) header format.

## Behaviour

- **New root trace** — when no `Traceparent` header is present, a new 128-bit
  trace ID and 64-bit span ID are generated and injected into the request.
- **Child span** — when an incoming `Traceparent` header is detected the trace
  ID is preserved, the incoming span ID is recorded as the parent, and a fresh
  child span ID is generated.
- The resolved `Traceparent` value is:
  - Set on the outgoing (upstream) request so the backend can continue the trace.
  - Echoed back in the response headers so API clients can correlate responses.

## Context Access

Downstream handlers can retrieve trace information via `TraceFromContext`:

```go
info, ok := middleware.TraceFromContext(r.Context())
if ok {
    log.Printf("trace=%s span=%s", info.TraceID, info.SpanID)
}
```

## Usage

```go
mux := http.NewServeMux()
handler := middleware.Tracing(mux)
```

## Header Format

```
Traceparent: 00-<traceID>-<spanID>-01
```

| Field    | Length  | Description                          |
|----------|---------|--------------------------------------|
| version  | 2 hex   | Always `00`                          |
| traceID  | 32 hex  | Unique identifier for the trace      |
| spanID   | 16 hex  | Identifier for this hop              |
| flags    | 2 hex   | `01` = sampled                       |
