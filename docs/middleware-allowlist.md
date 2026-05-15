# AllowList Middleware

The `AllowList` middleware restricts inbound requests to a configured set of
paths. Requests to any path not on the list are rejected with a configurable
HTTP status code (default **403 Forbidden**).

## Configuration

| Field | Type | Default | Description |
|---|---|---|---|
| `Paths` | `[]string` | `[]` | Permitted paths. Empty = allow all. |
| `PrefixMatch` | `bool` | `false` | Match by prefix instead of exact path. |
| `DeniedStatus` | `int` | `403` | Status code returned for blocked requests. |
| `DeniedMessage` | `string` | `{"error":"path not allowed"}` | Response body for blocked requests. |

## Usage

```go
cfg := middleware.DefaultAllowListConfig()
cfg.Paths = []string{"/api/", "/health"}
cfg.PrefixMatch = true

handler := middleware.AllowList(cfg)(upstreamHandler)
```

## Behaviour

- When `Paths` is empty the middleware is a **no-op** and all requests pass.
- With `PrefixMatch = false` (default) the full request path must equal one
  of the configured paths exactly.
- With `PrefixMatch = true` the request path only needs to start with one of
  the configured entries.
- Blocked responses always include `Content-Type: application/json`.

## Chaining

Place `AllowList` **early** in the middleware chain — before auth, rate-limiting,
or logging — so that disallowed paths are rejected with minimal overhead.

```go
chain := middleware.RequestID(
    middleware.AllowList(cfg)(
        middleware.RequireAuth(authCfg)(proxyHandler),
    ),
)
```
