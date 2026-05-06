package middleware

import (
	"context"
	"net/http"
	"time"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	// Timeout is the maximum duration for a proxied request.
	Timeout time.Duration
	// Message is the response body sent when the deadline is exceeded.
	Message string
}

// DefaultTimeoutConfig returns a TimeoutConfig with sensible defaults.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout: 30 * time.Second,
		Message: "gateway timeout",
	}
}

// RequestTimeout wraps h and cancels the request context after cfg.Timeout.
// If the handler does not finish in time, a 504 Gateway Timeout is returned.
func RequestTimeout(cfg TimeoutConfig) func(http.Handler) http.Handler {
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeoutConfig().Timeout
	}
	msg := cfg.Message
	if msg == "" {
		msg = DefaultTimeoutConfig().Message
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), cfg.Timeout)
			defer cancel()

			done := make(chan struct{})
			pw := &panicWriter{ResponseWriter: w}

			go func() {
				defer close(done)
				next.ServeHTTP(pw, r.WithContext(ctx))
			}()

			select {
			case <-done:
				// handler completed in time
			case <-ctx.Done():
				if !pw.written {
					w.WriteHeader(http.StatusGatewayTimeout)
					_, _ = w.Write([]byte(msg))
				}
			}
		})
	}
}

// panicWriter tracks whether the underlying ResponseWriter has been written to.
type panicWriter struct {
	http.ResponseWriter
	written bool
}

func (pw *panicWriter) WriteHeader(code int) {
	pw.written = true
	pw.ResponseWriter.WriteHeader(code)
}

func (pw *panicWriter) Write(b []byte) (int, error) {
	pw.written = true
	return pw.ResponseWriter.Write(b)
}
