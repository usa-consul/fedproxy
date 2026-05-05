package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// ResponseRecorder wraps http.ResponseWriter to capture the status code.
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Written    int64
}

func (r *ResponseRecorder) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.Written += int64(n)
	return n, err
}

// NewResponseRecorder returns a ResponseRecorder with a default 200 status.
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w, StatusCode: http.StatusOK}
}

// RequestLogger returns an HTTP middleware that logs each request using slog.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := NewResponseRecorder(w)

			next.ServeHTTP(rec, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"status", rec.StatusCode,
				"bytes", rec.Written,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
