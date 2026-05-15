# RewriteHost Middleware

`RewriteHost` rewrites the HTTP `Host` header before the request reaches the
upstream.  This is useful when the public-facing hostname differs from the
internal upstream hostname, or when a single proxy fronts multiple upstream
services distinguished by host.

The original `Host` value is preserved in the `X-Forwarded-Host` request
header so that upstream services can reconstruct the original URL if needed.

## Configuration

```go
type RewriteHostConfig struct {
    StaticHost  string            // Replace Host with this fixed value (highest priority)
    Rules       []HostRewriteRule // Ordered list of from→to mappings
    PassThrough bool              // Document intent; Host unchanged when no rule matches
}

type HostRewriteRule struct {
    From string // Exact incoming Host value (case-insensitive)
    To   string // Replacement Host value
}
```

### Priority

1. **StaticHost** — applied to every request when non-empty.
2. **Rules** — evaluated in order; the first case-insensitive match wins.
3. If nothing matches, the `Host` header is left unchanged.

## Usage

```go
cfg := middleware.RewriteHostConfig{
    Rules: []middleware.HostRewriteRule{
        {From: "api.example.gov", To: "api-internal.cluster.local"},
        {From: "app.example.gov", To: "app-internal.cluster.local"},
    },
}

handler := middleware.RewriteHost(cfg)(proxyHandler)
```

### Static upstream

```go
cfg := middleware.RewriteHostConfig{
    StaticHost: "backend.internal:8080",
}
```

## Headers

| Header | Direction | Description |
|---|---|---|
| `Host` | Request | Rewritten to the resolved target host |
| `X-Forwarded-Host` | Request | Set to the original incoming `Host` value |

## Notes

- Host matching is **case-insensitive** and must be an **exact** match
  (including port when present, e.g. `old.example.com:8443`).
- `X-Forwarded-Host` is only added when the host is actually rewritten.
- Compatible with all other fedproxy middleware; place early in the chain so
  downstream middleware and the proxy see the rewritten host.
