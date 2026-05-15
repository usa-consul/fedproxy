package middleware

import (
	"net/http"
	"strings"
)

// RedirectHTTPSConfig holds configuration for the HTTPS redirect middleware.
type RedirectHTTPSConfig struct {
	// Enabled controls whether HTTP requests are redirected to HTTPS.
	Enabled bool

	// StatusCode is the HTTP status code used for the redirect.
	// Defaults to 301 Moved Permanently.
	StatusCode int

	// ExemptPaths lists paths that bypass the redirect (e.g. health checks).
	ExemptPaths []string
}

// DefaultRedirectHTTPSConfig returns a sensible default configuration.
func DefaultRedirectHTTPSConfig() RedirectHTTPSConfig {
	return RedirectHTTPSConfig{
		Enabled:    true,
		StatusCode: http.StatusMovedPermanently,
		ExemptPaths: []string{"/healthz"},
	}
}

// RedirectHTTPS returns middleware that redirects plain HTTP requests to HTTPS.
// Requests already arriving over TLS (or with X-Forwarded-Proto: https) pass through.
func RedirectHTTPS(cfg RedirectHTTPSConfig) func(http.Handler) http.Handler {
	if cfg.StatusCode == 0 {
		cfg.StatusCode = http.StatusMovedPermanently
	}

	exempt := buildExemptSet(cfg.ExemptPaths)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := exempt[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			// Already HTTPS — check TLS state and forwarded proto header.
			if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
				next.ServeHTTP(w, r)
				return
			}

			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, cfg.StatusCode)
		})
	}
}
