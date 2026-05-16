package middleware

import (
	"log"
	"net/http"
	"time"
)

// RequestLogConfig controls structured per-request audit logging.
type RequestLogConfig struct {
	// Logger is the destination logger; defaults to the standard logger.
	Logger *log.Logger
	// SkipPaths are exact paths that will not be logged (e.g. health checks).
	SkipPaths []string
	// IncludeRequestID attaches the X-Request-ID value when present.
	IncludeRequestID bool
	// IncludeTraceID attaches the traceparent trace-id when present.
	IncludeTraceID bool
}

// DefaultRequestLogConfig returns a sensible baseline configuration.
func DefaultRequestLogConfig() RequestLogConfig {
	return RequestLogConfig{
		Logger:           log.Default(),
		IncludeRequestID: true,
		IncludeTraceID:   true,
	}
}

// RequestAuditLog emits a structured log line for every proxied request.
// It is distinct from AccessLog in that it surfaces fedproxy-specific context
// (request-id, trace-id) and is intended for audit / compliance pipelines.
func RequestAuditLog(cfg RequestLogConfig) func(http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	skip := buildExemptSet(cfg.SkipPaths)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, exempt := skip[r.URL.Path]; exempt {
				next.ServeHTTP(w, r)
				return
			}

			rec := NewResponseRecorder(w)
			start := time.Now()

			next.ServeHTTP(rec, r)

			duration := time.Since(start)

			fields := []interface{}{
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.Status(),
				"bytes", rec.BytesWritten(),
				"duration_ms", duration.Milliseconds(),
				"remote", r.RemoteAddr,
			}

			if cfg.IncludeRequestID {
				if id := r.Header.Get("X-Request-ID"); id != "" {
					fields = append(fields, "request_id", id)
				} else if id = FromContext(r.Context()); id != "" {
					fields = append(fields, "request_id", id)
				}
			}

			if cfg.IncludeTraceID {
				if ti := TraceFromContext(r.Context()); ti.TraceID != "" {
					fields = append(fields, "trace_id", ti.TraceID)
				}
			}

			cfg.Logger.Println(fields...)
		})
	}
}
