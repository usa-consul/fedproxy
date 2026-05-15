package middleware

import (
	"net/http"
	"strings"
)

// AllowListConfig controls path-based allowlisting.
type AllowListConfig struct {
	// Paths is the list of exact paths or prefixes that are permitted.
	// If empty, all paths are allowed.
	Paths []string
	// PrefixMatch enables prefix matching instead of exact matching.
	PrefixMatch bool
	// DeniedStatus is the HTTP status returned for blocked requests (default 403).
	DeniedStatus int
	// DeniedMessage is the JSON body returned for blocked requests.
	DeniedMessage string
}

// DefaultAllowListConfig returns a permissive default configuration.
func DefaultAllowListConfig() AllowListConfig {
	return AllowListConfig{
		Paths:         []string{},
		PrefixMatch:   false,
		DeniedStatus:  http.StatusForbidden,
		DeniedMessage: `{"error":"path not allowed"}`,
	}
}

// AllowList restricts inbound requests to a configured set of paths.
// When Paths is empty the middleware is a no-op.
func AllowList(cfg AllowListConfig) func(http.Handler) http.Handler {
	if cfg.DeniedStatus == 0 {
		cfg.DeniedStatus = http.StatusForbidden
	}
	if cfg.DeniedMessage == "" {
		cfg.DeniedMessage = `{"error":"path not allowed"}`
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(cfg.Paths) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			if isAllowed(r.URL.Path, cfg.Paths, cfg.PrefixMatch) {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(cfg.DeniedStatus)
			_, _ = w.Write([]byte(cfg.DeniedMessage))
		})
	}
}

func isAllowed(path string, allowed []string, prefix bool) bool {
	for _, p := range allowed {
		if prefix {
			if strings.HasPrefix(path, p) {
				return true
			}
		} else {
			if path == p {
				return true
			}
		}
	}
	return false
}
