package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter tracks request counts per IP using a sliding window.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	max      int
	window   time.Duration
}

type bucket struct {
	count     int
	windowEnd time.Time
}

// NewRateLimiter creates a RateLimiter allowing max requests per window duration.
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		max:     max,
		window:  window,
	}
}

// Allow returns true if the given key is within the rate limit.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[key]
	if !ok || now.After(b.windowEnd) {
		r.buckets[key] = &bucket{count: 1, windowEnd: now.Add(r.window)}
		return true
	}
	if b.count >= r.max {
		return false
	}
	b.count++
	return true
}

// RateLimit returns middleware that enforces per-IP rate limiting.
func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !limiter.Allow(ip) {
				http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the remote IP, stripping the port.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	host := r.RemoteAddr
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}
