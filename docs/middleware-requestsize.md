# RequestSize Middleware

The `RequestSize` middleware enforces per-dimension size limits on every
incoming HTTP request before it is forwarded to the upstream service.

## Why it matters

Federal proxy deployments often sit at the edge of a network and must
defend upstream services against abnormally large requests that could
cause memory exhaustion or slow-loris-style denial-of-service attacks.

## Limits enforced

| Dimension | Default | Rejection code |
|-----------|---------|----------------|
| Request URI length | 4 KB | `414 URI Too Long` |
| Combined header size | 8 KB | `431 Request Header Fields Too Large` |
| Request body | 1 MB | `413 Request Entity Too Large` |

Setting any limit to `0` disables that specific check.

## Usage

```go
import "github.com/example/fedproxy/internal/middleware"

// Use defaults
cfg := middleware.DefaultRequestSizeConfig()

// Or customise
cfg := middleware.RequestSizeConfig{
    MaxHeaderBytes: 16 * 1024,        // 16 KB headers
    MaxURIBytes:    2 * 1024,         // 2 KB URI
    MaxBodyBytes:   10 * 1024 * 1024, // 10 MB body
}

handler := middleware.RequestSize(cfg)(nextHandler)
```

## Notes

- The body limit uses `http.MaxBytesReader` so the connection is closed
  cleanly after the limit is exceeded during streaming reads.
- The URI check compares `r.RequestURI` (the raw, unparsed form) so
  percent-encoded characters count toward the limit.
- Header size is computed as the sum of all header name and value byte
  lengths; HTTP/2 pseudo-headers are not included.
