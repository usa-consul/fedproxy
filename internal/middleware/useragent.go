package middleware

import (
	"net/http"
	"strings"
)

// DefaultUserAgentConfig returns a config that blocks no user agents.
func DefaultUserAgentConfig() UserAgentConfig {
	return UserAgentConfig{
		BlockedAgents: []string{},
		RequireNonEmpty: false,
	}
}

// UserAgentConfig controls which User-Agent strings are blocked.
type UserAgentConfig struct {
	// BlockedAgents is a list of substrings; any UA containing one is rejected.
	BlockedAgents []string
	// RequireNonEmpty rejects requests that send no User-Agent header.
	RequireNonEmpty bool
}

// UserAgent blocks requests whose User-Agent header matches a blocked pattern
// or is absent when RequireNonEmpty is set.
func UserAgent(cfg UserAgentConfig, next http.Handler) http.Handler {
	lower := make([]string, len(cfg.BlockedAgents))
	for i, a := range cfg.BlockedAgents {
		lower[i] = strings.ToLower(a)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")

		if cfg.RequireNonEmpty && strings.TrimSpace(ua) == "" {
			http.Error(w, `{"error":"User-Agent header required"}`, http.StatusBadRequest)
			return
		}

		uaLower := strings.ToLower(ua)
		for _, blocked := range lower {
			if strings.Contains(uaLower, blocked) {
				http.Error(w, `{"error":"Forbidden User-Agent"}`, http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
