# IP Filter Middleware

The `IPFilter` middleware enforces IP-based access control using CIDR allow and block lists.

## Configuration

```go
type IPFilterConfig struct {
    AllowCIDRs []string // if non-empty, only these ranges are permitted
    BlockCIDRs []string // these ranges are always denied
    TrustProxy bool     // use X-Forwarded-For for client IP resolution
}
```

## Behavior

| Condition | Result |
|---|---|
| No rules configured | All IPs allowed |
| IP matches `BlockCIDRs` | `403 Forbidden` |
| `AllowCIDRs` set, IP not in list | `403 Forbidden` |
| `AllowCIDRs` set, IP matches | Passes through |
| IP in both allow and block | Block wins (`403`) |

## Usage

```go
cfg := middleware.IPFilterConfig{
    AllowCIDRs: []string{"10.0.0.0/8", "192.168.0.0/16"},
    BlockCIDRs: []string{"10.0.0.0/24"},
    TrustProxy: true,
}

mw, err := middleware.IPFilter(cfg)
if err != nil {
    log.Fatalf("invalid IP filter config: %v", err)
}

handler = mw(handler)
```

## Proxy Trust

When `TrustProxy` is `true`, the middleware reads the client IP from the
`X-Forwarded-For` header (first value) instead of `RemoteAddr`. Enable this
only when fedproxy sits behind a trusted load balancer or ingress controller.

## Federal / PIV Context

For agency deployments, IP filtering can be combined with SAML/PIV auth to
restrict access to known government network ranges (e.g., `.gov` data centers)
before SAML assertions are evaluated, providing a defense-in-depth posture.
