package middleware

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// MetricsSnapshot holds a point-in-time view of collected counters.
type MetricsSnapshot struct {
	TotalRequests  uint64        `json:"total_requests"`
	TotalErrors    uint64        `json:"total_errors"`
	TotalBytes     uint64        `json:"total_bytes_sent"`
	Uptime         string        `json:"uptime"`
	AvgLatencyMs   float64       `json:"avg_latency_ms"`
}

// metricsState holds global atomic counters for the metrics middleware.
var metricsState = struct {
	requests   uint64
	errors     uint64
	bytes      uint64
	latencySum uint64 // microseconds
	startTime  time.Time
}{
	startTime: time.Now(),
}

// DefaultMetricsConfig returns the default path for the metrics endpoint.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Path: "/_metrics",
	}
}

// MetricsConfig controls the metrics middleware behaviour.
type MetricsConfig struct {
	// Path is the URL path that exposes the JSON metrics snapshot.
	Path string
}

// Metrics is a middleware that records per-request counters and exposes them
// as a JSON endpoint at cfg.Path.
func Metrics(cfg MetricsConfig) func(http.Handler) http.Handler {
	if cfg.Path == "" {
		cfg = DefaultMetricsConfig()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == cfg.Path {
				serveMetrics(w, r)
				return
			}

			start := time.Now()
			rec := NewResponseRecorder(w)
			next.ServeHTTP(rec, r)

			atomic.AddUint64(&metricsState.requests, 1)
			atomic.AddUint64(&metricsState.bytes, uint64(rec.BytesWritten))
			atomic.AddUint64(&metricsState.latencySum, uint64(time.Since(start).Microseconds()))
			if rec.Status >= 500 {
				atomic.AddUint64(&metricsState.errors, 1)
			}
		})
	}
}

func serveMetrics(w http.ResponseWriter, _ *http.Request) {
	reqs := atomic.LoadUint64(&metricsState.requests)
	latSum := atomic.LoadUint64(&metricsState.latencySum)

	var avgMs float64
	if reqs > 0 {
		avgMs = float64(latSum) / float64(reqs) / 1000.0
	}

	snap := MetricsSnapshot{
		TotalRequests: reqs,
		TotalErrors:   atomic.LoadUint64(&metricsState.errors),
		TotalBytes:    atomic.LoadUint64(&metricsState.bytes),
		Uptime:        time.Since(metricsState.startTime).Round(time.Second).String(),
		AvgLatencyMs:  avgMs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(snap)
}
