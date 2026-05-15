package middleware

import (
	"net/http"
	"sync"
	"time"
)

// DefaultDedupeConfig returns a DedupeConfig with sensible defaults.
func DefaultDedupeConfig() DedupeConfig {
	return DedupeConfig{
		TTL:    500 * time.Millisecond,
		Header: "X-Idempotency-Key",
	}
}

// DedupeConfig controls idempotency-key-based request deduplication.
type DedupeConfig struct {
	// TTL is how long a completed response is cached for replay.
	TTL time.Duration
	// Header is the request header carrying the idempotency key.
	Header string
}

type cachedResponse struct {
	status  int
	headers http.Header
	body    []byte
	expires time.Time
}

// Dedupe returns middleware that replays the first response for any request
// that carries an idempotency key already seen within the TTL window.
func Dedupe(cfg DedupeConfig) func(http.Handler) http.Handler {
	if cfg.TTL <= 0 {
		cfg.TTL = DefaultDedupeConfig().TTL
	}
	if cfg.Header == "" {
		cfg.Header = DefaultDedupeConfig().Header
	}

	var mu sync.Mutex
	cache := map[string]*cachedResponse{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get(cfg.Header)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			mu.Lock()
			if entry, ok := cache[key]; ok && time.Now().Before(entry.expires) {
				mu.Unlock()
				for k, vals := range entry.headers {
					for _, v := range vals {
						w.Header().Add(k, v)
					}
				}
				w.Header().Set("X-Dedupe-Replay", "true")
				w.WriteHeader(entry.status)
				w.Write(entry.body) //nolint:errcheck
				return
			}
			mu.Unlock()

			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)

			mu.Lock()
			cache[key] = &cachedResponse{
				status:  rec.Status(),
				headers: w.Header().Clone(),
				body:    rec.Body(),
				expires: time.Now().Add(cfg.TTL),
			}
			mu.Unlock()
		})
	}
}
