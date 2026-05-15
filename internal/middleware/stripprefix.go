package middleware

import (
	"net/http"
	"strings"
)

// StripPrefixConfig holds configuration for the StripPrefix middleware.
type StripPrefixConfig struct {
	// Prefixes is the list of path prefixes to strip before forwarding.
	Prefixes []string
}

// DefaultStripPrefixConfig returns a StripPrefixConfig with no prefixes.
func DefaultStripPrefixConfig() StripPrefixConfig {
	return StripPrefixConfig{}
}

// StripPrefix removes the first matching prefix from the request path before
// passing the request to the next handler. If no prefix matches the request
// is forwarded unchanged. A stripped path that becomes empty is normalised
// to "/". The original path is preserved in the X-Forwarded-Prefix header
// so upstream services can reconstruct absolute URLs when needed.
func StripPrefix(cfg StripPrefixConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, prefix := range cfg.Prefixes {
			if prefix == "" {
				continue
			}
			if strings.HasPrefix(r.URL.Path, prefix) {
				stripped := strings.TrimPrefix(r.URL.Path, prefix)
				if stripped == "" {
					stripped = "/"
				}
				// Clone the request so we do not mutate the original.
				mod := r.Clone(r.Context())
				mod.Header.Set("X-Forwarded-Prefix", prefix)
				mod.URL.Path = stripped
				if mod.URL.RawPath != "" {
					rawStripped := strings.TrimPrefix(mod.URL.RawPath, prefix)
					if rawStripped == "" {
						rawStripped = "/"
					}
					mod.URL.RawPath = rawStripped
				}
				next.ServeHTTP(w, mod)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
