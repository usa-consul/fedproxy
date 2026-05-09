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

// applyRule rewrites the given path according to rule, returning the rewritten
// path and true if the rule matched, or the original path and false otherwise.
func applyRule(path string, rule RewriteRule) (string, bool) {
	if rule.StripPrefix == "" || !strings.HasPrefix(path, rule.StripPrefix) {
		return path, false
	}
	trimmed := strings.TrimPrefix(path, rule.StripPrefix)
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return rule.AddPrefix + trimmed, true
}

// PathRewrite returns middleware that rewrites request URL paths according to
// the provided rules. Rules are evaluated in order; the first matching rule
// is applied and evaluation stops.
func PathRewrite(cfg RewriteConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			for _, rule := range cfg.Rules {
				if newPath, ok := applyRule(path, rule); ok {
					r = r.Clone(r.Context())
					r.URL.Path = newPath
					if r.URL.RawPath != "" {
						if newRawPath, rawOk := applyRule(r.URL.RawPath, rule); rawOk {
							r.URL.RawPath = newRawPath
						}
					}
					break
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
