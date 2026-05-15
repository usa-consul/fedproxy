package middleware

import (
	"net/http"
	"sync"
	"time"
)

// ThrottleConfig controls concurrency-based throttling.
type ThrottleConfig struct {
	// MaxConcurrent is the maximum number of requests handled simultaneously.
	MaxConcurrent int
	// QueueTimeout is how long a request waits for a slot before 503.
	QueueTimeout time.Duration
	// StatusCode is returned when the queue is full (default 503).
	StatusCode int
}

// DefaultThrottleConfig returns sensible defaults.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		MaxConcurrent: 100,
		QueueTimeout:  5 * time.Second,
		StatusCode:    http.StatusServiceUnavailable,
	}
}

type throttler struct {
	cfg  ThrottleConfig
	sem  chan struct{}
	once sync.Once
}

func (t *throttler) init() {
	t.once.Do(func() {
		if t.sem == nil {
			t.sem = make(chan struct{}, t.cfg.MaxConcurrent)
		}
	})
}

// Throttle limits the number of concurrently processed requests.
// Requests that cannot acquire a slot within QueueTimeout receive
// a configurable error status (default 503).
func Throttle(cfg ThrottleConfig, next http.Handler) http.Handler {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = DefaultThrottleConfig().MaxConcurrent
	}
	if cfg.QueueTimeout <= 0 {
		cfg.QueueTimeout = DefaultThrottleConfig().QueueTimeout
	}
	if cfg.StatusCode == 0 {
		cfg.StatusCode = DefaultThrottleConfig().StatusCode
	}

	t := &throttler{
		cfg: cfg,
		sem: make(chan struct{}, cfg.MaxConcurrent),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case t.sem <- struct{}{}:
			// slot acquired
		case <-time.After(cfg.QueueTimeout):
			http.Error(w, http.StatusText(cfg.StatusCode), cfg.StatusCode)
			return
		}
		defer func() { <-t.sem }()
		next.ServeHTTP(w, r)
	})
}
