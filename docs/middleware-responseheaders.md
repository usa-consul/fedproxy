# ResponseHeaders Middleware

`ResponseHeaders` mutates HTTP response headers produced by the upstream before
they reach the client. It supports three independent operations that are applied
in order: **Remove → Set → Add**.

## Configuration

```go
type ResponseHeadersConfig struct {
    Add    map[string]string // append without replacing existing values
    Set    map[string]string // overwrite (or create) unconditionally
    Remove []string          // strip from upstream response
}
```

## Usage

```go
cfg := middleware.ResponseHeadersConfig{
    Remove: []string{"Server", "X-Powered-By"},
    Set: map[string]string{
        "Server": "fedproxy/1.0",
    },
    Add: map[string]string{
        "X-Content-Type-Options": "nosniff",
    },
}

handler := middleware.ResponseHeaders(cfg)(nextHandler)
```

## Operation order

| Step | Action |
|------|--------|
| 1    | Headers listed in `Remove` are deleted from the upstream response. |
| 2    | Headers in `Set` overwrite any existing value (including ones just removed). |
| 3    | Headers in `Add` are appended, preserving pre-existing values. |

Mutations are applied lazily — immediately before the status code is flushed to
the client — so upstream handlers may still modify headers freely up to that
point.

## Defaults

`DefaultResponseHeadersConfig()` returns an empty configuration that is a
complete no-op; no headers are added, removed, or changed.

## Notes

- Header names are case-insensitive (delegated to `net/http`).
- `Remove` normalises names to lowercase at construction time to avoid
  repeated allocations per request.
- Compatible with all other fedproxy middleware; place it after the upstream
  proxy handler in the chain so upstream headers are already populated.
