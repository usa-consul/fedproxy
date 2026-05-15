package middleware

import (
	"context"
	"net/http"
	"strings"
)

// requestTagKey is the context key for request tags.
type requestTagKey struct{}

// DefaultRequestTagConfig returns a RequestTagConfig with sensible defaults.
func DefaultRequestTagConfig() RequestTagConfig {
	return RequestTagConfig{
		Header:    "X-Request-Tag",
		MaxLength: 64,
		Allowed:   nil, // nil means all values accepted
	}
}

// RequestTagConfig controls the RequestTag middleware.
type RequestTagConfig struct {
	// Header is the request header to read the tag from.
	Header string
	// MaxLength is the maximum allowed tag length. Longer values are truncated.
	MaxLength int
	// Allowed is an optional allowlist of accepted tag values.
	// When non-nil, requests with tags not in the set are silently cleared.
	Allowed []string
}

// TagFromContext retrieves the request tag stored in ctx, or an empty string.
func TagFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestTagKey{}).(string)
	return v
}

// RequestTag reads an optional caller-supplied tag from a request header,
// sanitises it, and stores it in the request context so downstream handlers
// and loggers can attach it to their output.
func RequestTag(cfg RequestTagConfig) func(http.Handler) http.Handler {
	if cfg.Header == "" {
		cfg.Header = "X-Request-Tag"
	}
	if cfg.MaxLength <= 0 {
		cfg.MaxLength = 64
	}

	allowSet := make(map[string]struct{}, len(cfg.Allowed))
	for _, v := range cfg.Allowed {
		allowSet[strings.ToLower(v)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tag := strings.TrimSpace(r.Header.Get(cfg.Header))

			if len(tag) > cfg.MaxLength {
				tag = tag[:cfg.MaxLength]
			}

			if len(allowSet) > 0 && tag != "" {
				if _, ok := allowSet[strings.ToLower(tag)]; !ok {
					tag = ""
				}
			}

			if tag != "" {
				r = r.WithContext(context.WithValue(r.Context(), requestTagKey{}, tag))
				w.Header().Set("X-Request-Tag", tag)
			}

			next.ServeHTTP(w, r)
		})
	}
}
