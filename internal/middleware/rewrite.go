package middleware

import (
	"net/http"
	"strings"
)

// RewriteRule defines a path prefix rewrite: strip StripPrefix, then prepend AddPrefix.
type RewriteRule struct {
	StripPrefix string
	AddPrefix   string
}

// RewriteConfig holds configuration for the PathRewrite middleware.
type RewriteConfig struct {
	Rules []RewriteRule
}

// DefaultRewriteConfig returns an empty RewriteConfig (no-op).
func DefaultRewriteConfig() RewriteConfig {
	return RewriteConfig{}
}

// PathRewrite returns middleware that rewrites request URL paths according to
// the provided rules. Rules are evaluated in order; the first matching rule
// is applied and evaluation stops.
func PathRewrite(cfg RewriteConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			for _, rule := range cfg.Rules {
				if rule.StripPrefix != "" && strings.HasPrefix(path, rule.StripPrefix) {
					trimmed := strings.TrimPrefix(path, rule.StripPrefix)
					if !strings.HasPrefix(trimmed, "/") {
						trimmed = "/" + trimmed
					}
					r = r.Clone(r.Context())
					r.URL.Path = rule.AddPrefix + trimmed
					if r.URL.RawPath != "" {
						rawTrimmed := strings.TrimPrefix(r.URL.RawPath, rule.StripPrefix)
						if !strings.HasPrefix(rawTrimmed, "/") {
							rawTrimmed = "/" + rawTrimmed
						}
						r.URL.RawPath = rule.AddPrefix + rawTrimmed
					}
					break
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
