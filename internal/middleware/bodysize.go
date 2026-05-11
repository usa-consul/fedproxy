package middleware

import (
	"fmt"
	"net/http"
)

// DefaultBodySizeConfig returns a BodySizeConfig with a 1 MB limit.
func DefaultBodySizeConfig() BodySizeConfig {
	return BodySizeConfig{
		MaxBytes: 1 << 20, // 1 MB
	}
}

// BodySizeConfig configures the maximum allowed request body size.
type BodySizeConfig struct {
	// MaxBytes is the maximum number of bytes allowed in the request body.
	// A value of 0 disables the limit.
	MaxBytes int64
}

// LimitBody returns middleware that rejects requests whose Content-Length
// exceeds cfg.MaxBytes, and wraps the body reader so that reads beyond the
// limit also result in a 413 response. If MaxBytes is 0 the middleware is a
// no-op pass-through.
func LimitBody(cfg BodySizeConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.MaxBytes <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Reject early when Content-Length is known and already over limit.
			if r.ContentLength > cfg.MaxBytes {
				http.Error(
					w,
					fmt.Sprintf("request body too large (limit %d bytes)", cfg.MaxBytes),
					http.StatusRequestEntityTooLarge,
				)
				return
			}

			// Wrap the body so that streaming reads are also capped.
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBytes)
			}

			next.ServeHTTP(w, r)
		})
	}
}
