# JWT Middleware

The `JWT` middleware validates HS256-signed JSON Web Tokens on incoming requests,
storing parsed claims in the request context for downstream handlers.

## Configuration

```go
cfg := middleware.DefaultJWTConfig("your-secret-key")
// Optionally override:
cfg.ExemptPaths  = []string{"/health", "/metrics", "/login"}
cfg.ClockSkewSec = 60  // tolerate up to 60 s of clock skew
cfg.HeaderName   = "Authorization"  // default
```

## Usage

```go
handler := middleware.JWT(cfg, nextHandler)
```

The middleware reads the `Authorization: Bearer <token>` header, verifies the
HMAC-SHA256 signature, checks the `exp` claim (with optional clock skew), and
rejects requests with a `401 Unauthorized` response if validation fails.

Paths listed in `ExemptPaths` bypass token validation entirely — useful for
health checks and login endpoints.

## Accessing Claims

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.ClaimsFromContext(r.Context())
    if ok {
        fmt.Println("subject:", claims.Subject)
        fmt.Println("issuer:",  claims.Issuer)
    }
}
```

## Error Responses

| Condition              | Status |
|------------------------|--------|
| Missing token          | 401    |
| Malformed token        | 401    |
| Signature mismatch     | 401    |
| Token expired          | 401    |

## Integration with Other Middleware

The JWT middleware is designed to be composed with the rest of the fedproxy
stack. A typical chain for a protected API route:

```go
http.Handle("/api/",
    middleware.RequestID(
        middleware.NewRateLimiter(rlCfg).RateLimit(
            middleware.JWT(jwtCfg,
                middleware.Tracing(
                    proxyHandler,
                ),
            ),
        ),
    ),
)
```
