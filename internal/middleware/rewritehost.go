package middleware

import (
	"net/http"
	"strings"
)

// RewriteHostConfig controls how the Host header is rewritten before
// the request is forwarded to the upstream.
type RewriteHostConfig struct {
	// StaticHost replaces the Host header with a fixed value when non-empty.
	StaticHost string

	// Rules maps incoming host values (or prefixes when UsePrefix is true)
	// to replacement host values.  Rules are evaluated in declaration order;
	// the first match wins.
	Rules []HostRewriteRule

	// PassThrough, when true, leaves the Host header unchanged when no rule
	// matches.  When false (the default) and StaticHost is also empty, the
	// Host is left unchanged as well — this field mainly serves as
	// documentation of intent.
	PassThrough bool
}

// HostRewriteRule pairs an incoming host pattern with its replacement.
type HostRewriteRule struct {
	// From is the exact incoming Host value to match (port included if
	// present, e.g. "old.example.com:8080").
	From string
	// To is the replacement Host value.
	To string
}

// DefaultRewriteHostConfig returns a no-op configuration.
func DefaultRewriteHostConfig() RewriteHostConfig {
	return RewriteHostConfig{PassThrough: true}
}

// RewriteHost returns middleware that rewrites the incoming Host header
// before passing the request to the next handler.  It also sets
// X-Forwarded-Host to the original value so that upstream services can
// reconstruct the original URL when needed.
func RewriteHost(cfg RewriteHostConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			original := r.Host
			if original == "" {
				original = r.Header.Get("Host")
			}

			newHost := resolve(cfg, original)
			if newHost != "" && newHost != original {
				if original != "" {
					r.Header.Set("X-Forwarded-Host", original)
				}
				r = r.Clone(r.Context())
				r.Host = newHost
				r.Header.Set("Host", newHost)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// resolve picks the replacement host according to cfg priority:
//  1. StaticHost (if set)
//  2. First matching rule
//  3. Empty string (caller keeps original)
func resolve(cfg RewriteHostConfig, host string) string {
	if cfg.StaticHost != "" {
		return cfg.StaticHost
	}
	for _, rule := range cfg.Rules {
		if strings.EqualFold(rule.From, host) {
			return rule.To
		}
	}
	return ""
}
