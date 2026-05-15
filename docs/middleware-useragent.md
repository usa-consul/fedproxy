# UserAgent Middleware

The `UserAgent` middleware inspects the incoming `User-Agent` header and blocks
requests that match a configured deny-list or that omit the header entirely.

## Configuration

```go
type UserAgentConfig struct {
    // BlockedAgents is a list of case-insensitive substrings.
    // A request whose User-Agent contains any entry is rejected with 403.
    BlockedAgents []string

    // RequireNonEmpty rejects requests with an absent or blank
    // User-Agent header with 400 Bad Request.
    RequireNonEmpty bool
}
```

`DefaultUserAgentConfig()` returns an empty allow-all configuration.

## Usage

```go
cfg := middleware.UserAgentConfig{
    BlockedAgents:   []string{"curl", "sqlmap", "nikto", "masscan"},
    RequireNonEmpty: true,
}

handler := middleware.UserAgent(cfg, nextHandler)
```

## Behaviour

| Condition | Response |
|-----------|----------|
| `User-Agent` absent and `RequireNonEmpty: true` | `400 Bad Request` |
| `User-Agent` matches a blocked substring (case-insensitive) | `403 Forbidden` |
| All checks pass | Request forwarded to next handler |

## Notes

- Matching is substring-based and case-insensitive, so `"curl"` blocks
  `curl/7.68.0`, `LibCURL/8.0`, etc.
- Place this middleware early in the chain, before auth and rate-limiting, to
  shed unwanted traffic as cheaply as possible.
- Responses are plain JSON strings for consistency with other fedproxy error
  responses.
