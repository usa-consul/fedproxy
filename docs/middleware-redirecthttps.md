# RedirectHTTPS Middleware

The `RedirectHTTPS` middleware enforces HTTPS by redirecting any plain HTTP
request to the equivalent `https://` URL. It is safe to use behind a TLS-
terminating load balancer that sets the `X-Forwarded-Proto` header.

## Configuration

```go
type RedirectHTTPSConfig struct {
    Enabled     bool
    StatusCode  int      // default: 301
    ExemptPaths []string // default: ["/healthz"]
}
```

| Field | Default | Description |
|---|---|---|
| `Enabled` | `true` | Toggle the redirect on/off. |
| `StatusCode` | `301` | HTTP status used for the redirect (`301` or `307`). |
| `ExemptPaths` | `["/healthz"]` | Paths that bypass the redirect. |

## Behaviour

1. If `Enabled` is `false`, all requests pass through unchanged.
2. Requests whose path matches an `ExemptPaths` entry are passed through.
3. Requests that already carry `X-Forwarded-Proto: https` or arrive over a
   native TLS connection (`r.TLS != nil`) are passed through.
4. All other requests receive a redirect to `https://<host><path>`.

## Usage

```go
cfg := middleware.DefaultRedirectHTTPSConfig()
// Optionally use 307 so browsers re-POST correctly:
cfg.StatusCode = http.StatusTemporaryRedirect

handler := middleware.RedirectHTTPS(cfg)(myHandler)
```

## Notes

- Place this middleware early in the chain, before auth or rate-limiting, so
  unauthenticated HTTP probes are redirected cheaply.
- Health-check paths are exempt by default so that load-balancer probes over
  plain HTTP continue to work without special firewall rules.
