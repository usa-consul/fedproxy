package middleware

import (
	"fmt"
	"net/http"
)

// DefaultRequestSizeConfig returns a conservative default limiting requests to 1 MB.
func DefaultRequestSizeConfig() RequestSizeConfig {
	return RequestSizeConfig{
		MaxHeaderBytes: 8 * 1024,       // 8 KB
		MaxURIBytes:    4 * 1024,       // 4 KB
		MaxBodyBytes:   1 * 1024 * 1024, // 1 MB
	}
}

// RequestSizeConfig controls per-dimension size limits for incoming requests.
type RequestSizeConfig struct {
	// MaxHeaderBytes is the maximum size of all request headers combined (bytes).
	// Zero disables the check.
	MaxHeaderBytes int64

	// MaxURIBytes is the maximum length of the raw request URI (bytes).
	// Zero disables the check.
	MaxURIBytes int64

	// MaxBodyBytes is the maximum size of the request body (bytes).
	// Zero disables the check.
	MaxBodyBytes int64
}

// RequestSize enforces per-dimension size limits on incoming HTTP requests.
// It rejects oversized requests with 413 Request Entity Too Large or
// 431 Request Header Fields Too Large before the request reaches the proxy.
func RequestSize(cfg RequestSizeConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// --- URI length ---
			if cfg.MaxURIBytes > 0 && int64(len(r.RequestURI)) > cfg.MaxURIBytes {
				http.Error(w,
					fmt.Sprintf("URI too long: limit %d bytes", cfg.MaxURIBytes),
					http.StatusRequestURITooLong,
				)
				return
			}

			// --- Header size ---
			if cfg.MaxHeaderBytes > 0 {
				var total int64
				for name, vals := range r.Header {
					total += int64(len(name))
					for _, v := range vals {
						total += int64(len(v))
					}
				}
				if total > cfg.MaxHeaderBytes {
					http.Error(w,
						fmt.Sprintf("request headers too large: limit %d bytes", cfg.MaxHeaderBytes),
						http.StatusRequestHeaderFieldsTooLarge,
					)
					return
				}
			}

			// --- Body size ---
			if cfg.MaxBodyBytes > 0 && r.Body != nil {
				// Content-Length fast path
				if r.ContentLength > cfg.MaxBodyBytes {
					http.Error(w,
						fmt.Sprintf("request body too large: limit %d bytes", cfg.MaxBodyBytes),
						http.StatusRequestEntityTooLarge,
					)
					return
				}
				r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodyBytes)
			}

			next.ServeHTTP(w, r)
		})
	}
}
