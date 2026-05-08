package middleware

import (
	"log"
	"net/http"
	"time"
)

// DefaultRetryConfig provides sensible defaults for the retry middleware.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	Delay:       100 * time.Millisecond,
	RetryOn:     []int{http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout},
}

// RetryConfig controls retry behaviour.
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	RetryOn     []int
	Logger      *log.Logger
}

type retryResponseWriter struct {
	http.ResponseWriter
	code int
	sniffed bool
}

func (r *retryResponseWriter) WriteHeader(code int) {
	r.code = code
	r.sniffed = true
}

func (r *retryResponseWriter) status() int {
	if !r.sniffed {
		return http.StatusOK
	}
	return r.code
}

// Retry wraps next and retries the request on configured status codes.
func Retry(cfg RetryConfig, next http.Handler) http.Handler {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = DefaultRetryConfig.MaxAttempts
	}
	if cfg.Delay <= 0 {
		cfg.Delay = DefaultRetryConfig.Delay
	}
	if len(cfg.RetryOn) == 0 {
		cfg.RetryOn = DefaultRetryConfig.RetryOn
	}
	retrySet := make(map[int]struct{}, len(cfg.RetryOn))
	for _, code := range cfg.RetryOn {
		retrySet[code] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
			rw := &retryResponseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)
			st := rw.status()
			if _, shouldRetry := retrySet[st]; !shouldRetry || attempt == cfg.MaxAttempts {
				if rw.sniffed {
					w.WriteHeader(st)
				}
				return
			}
			if cfg.Logger != nil {
				cfg.Logger.Printf("retry: attempt %d/%d got %d, retrying after %s",
					attempt, cfg.MaxAttempts, st, cfg.Delay)
			}
			time.Sleep(cfg.Delay)
		}
	})
}
