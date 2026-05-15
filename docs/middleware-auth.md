# Auth Middleware

The `RequireAuth` middleware enforces authentication on incoming requests based
on a configurable mode. It is designed to support federal agency authentication
flows including SAML assertions and PIV smart-card certificates.

## Modes

| Mode   | Description                                                      |
|--------|------------------------------------------------------------------|
| `none` | No authentication enforced. All requests pass through.           |
| `saml` | Requires an `X-Saml-Assertion` header to be present.            |
| `piv`  | Requires a verified TLS client certificate (PIV card).          |

## Configuration

```go
cfg := middleware.AuthConfig{
    Mode:        "saml",
    ExemptPaths: []string{"/health", "/_metrics", "/login"},
}

handler := middleware.RequireAuth(cfg, myHandler)
```

## Exempt Paths

Paths listed in `ExemptPaths` bypass all authentication checks. This is useful
for health-check and metrics endpoints that must remain accessible without
credentials.

## Default Config

```go
cfg := middleware.DefaultAuthConfig()
// Mode: "none", ExemptPaths: ["/health", "/_metrics"]
```

## PIV Notes

PIV mode inspects `r.TLS.PeerCertificates`. The upstream TLS listener must be
configured with `tls.RequireAnyClientCert` or `tls.RequireAndVerifyClientCert`
so that the certificate is available on the request.

## Response Codes

- `401 Unauthorized` — authentication missing or invalid.
- `200 OK` (or upstream response) — authentication satisfied.
