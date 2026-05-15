# BasicAuth Middleware

The `BasicAuth` middleware enforces HTTP Basic Authentication (RFC 7617) on incoming requests. It is intended for internal tooling, staging environments, or as a secondary auth layer behind SAML/PIV.

## Configuration

```go
cfg := middleware.DefaultBasicAuthConfig()
cfg.Credentials = map[string]string{
    "alice": os.Getenv("ALICE_PASSWORD"),
    "svcaccount": os.Getenv("SVC_PASSWORD"),
}
cfg.Realm = "my-agency-proxy"
cfg.ExemptPaths = []string{"/healthz", "/__metrics"}
```

| Field | Type | Default | Description |
|---|---|---|---|
| `Credentials` | `map[string]string` | `{}` | Username → password map. Populate from environment variables. |
| `Realm` | `string` | `"fedproxy"` | Realm string in `WWW-Authenticate` response header. |
| `ExemptPaths` | `[]string` | `["/healthz", "/__metrics"]` | Paths that skip authentication. |
| `Logger` | `*log.Logger` | `log.Default()` | Logger for auth failures. |

## Usage

```go
handler := middleware.BasicAuth(cfg, upstream)
http.ListenAndServe(addr, handler)
```

## Behavior

- Requests to exempt paths pass through without credentials.
- Missing or malformed `Authorization` header → `401 Unauthorized` with `WWW-Authenticate` challenge.
- Unknown username or wrong password → `401 Unauthorized`.
- Credential comparison uses `crypto/subtle.ConstantTimeCompare` to prevent timing attacks.
- All failures are logged with the remote address.

## Security Notes

- Store passwords in environment variables or a secrets manager — never in source code.
- Basic Auth transmits credentials in base64 (not encrypted); always run behind TLS.
- For stronger authentication, prefer the SAML or PIV flows built into fedproxy.
