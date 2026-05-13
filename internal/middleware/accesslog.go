package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// AccessLogConfig controls the format and destination of structured access logs.
type AccessLogConfig struct {
	// SkipPaths are request paths that will not be logged (e.g. health checks).
	SkipPaths []string
	// LogFunc receives the formatted log line. Defaults to fmt.Println.
	LogFunc func(line string)
}

// DefaultAccessLogConfig returns a sensible default configuration.
func DefaultAccessLogConfig() AccessLogConfig {
	return AccessLogConfig{
		SkipPaths: []string{"/healthz"},
		LogFunc:   func(line string) { fmt.Println(line) },
	}
}

// AccessLog is a structured access-log middleware that emits one log line per
// request in the format:
//
//	<method> <path> <status> <bytes>B <latency>ms <request-id> <trace-id>
func AccessLog(cfg AccessLogConfig) func(http.Handler) http.Handler {
	if cfg.LogFunc == nil {
		cfg.LogFunc = DefaultAccessLogConfig().LogFunc
	}

	skip := make(map[string]struct{}, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skip[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skipped := skip[r.URL.Path]; skipped {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)
			latency := time.Since(start).Milliseconds()

			reqID := FromContext(r.Context())
			traceID := TraceFromContext(r.Context())

			line := fmt.Sprintf("%s %s %d %dB %dms req=%s trace=%s",
				r.Method, r.URL.Path, rec.Status(), rec.BytesWritten(),
				latency, reqID, traceID,
			)
			cfg.LogFunc(line)
		})
	}
}
