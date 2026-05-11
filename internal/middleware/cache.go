package middleware

import (
	"net/http"
	"strconv"
	"time"
)

// CacheConfig controls response caching headers injected by the Cache middleware.
type CacheConfig struct {
	// MaxAge is the max-age value for cacheable responses. Zero disables caching.
	MaxAge time.Duration
	// CacheableStatuses lists HTTP status codes that are eligible for caching.
	CacheableStatuses []int
	// PrivatePaths are path prefixes that must never be publicly cached.
	PrivatePaths []string
}

// DefaultCacheConfig returns a conservative default suitable for a federal proxy.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxAge:            60 * time.Second,
		CacheableStatuses: []int{http.StatusOK, http.StatusNotModified},
		PrivatePaths:      []string{"/auth", "/saml", "/piv"},
	}
}

// Cache sets Cache-Control response headers based on the provided config.
// Paths matching PrivatePaths receive "no-store" regardless of other settings.
func Cache(cfg CacheConfig) func(http.Handler) http.Handler {
	privateSet := make(map[string]struct{}, len(cfg.PrivatePaths))
	for _, p := range cfg.PrivatePaths {
		privateSet[p] = struct{}{}
	}

	cacheableSet := make(map[int]struct{}, len(cfg.CacheableStatuses))
	for _, s := range cfg.CacheableStatuses {
		cacheableSet[s] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPrivatePath(r.URL.Path, cfg.PrivatePaths) {
				w.Header().Set("Cache-Control", "no-store")
				next.ServeHTTP(w, r)
				return
			}

			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)

			_, cacheable := cacheableSet[rec.Status()]
			if cfg.MaxAge > 0 && cacheable {
				secs := strconv.Itoa(int(cfg.MaxAge.Seconds()))
				w.Header().Set("Cache-Control", "public, max-age="+secs)
			} else {
				w.Header().Set("Cache-Control", "no-store")
			}
		})
	}
}

func isPrivatePath(reqPath string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if len(reqPath) >= len(prefix) && reqPath[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
