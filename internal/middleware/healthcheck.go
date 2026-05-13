package middleware

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthCheckConfig configures the health check endpoint.
type HealthCheckConfig struct {
	// Path is the URL path that serves the health response (default: /healthz).
	Path string
	// UpstreamURL is checked for liveness when provided.
	UpstreamURL string
	// Timeout for upstream probe (default: 3s).
	Timeout time.Duration
}

// DefaultHealthCheckConfig returns a config with sensible defaults.
func DefaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		Path:    "/healthz",
		Timeout: 3 * time.Second,
	}
}

type healthStatus struct {
	Status    string `json:"status"`
	Upstream  string `json:"upstream,omitempty"`
	Timestamp string `json:"timestamp"`
}

// requestCounter tracks total proxied requests for the health payload.
var requestCounter atomic.Int64

// IncrementRequestCount is called by other middleware to bump the counter.
func IncrementRequestCount() { requestCounter.Add(1) }

// HealthCheck intercepts requests to the configured path and returns a JSON
// health payload. All other requests are passed to next.
func HealthCheck(cfg HealthCheckConfig, next http.Handler) http.Handler {
	if cfg.Path == "" {
		cfg.Path = DefaultHealthCheckConfig().Path
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultHealthCheckConfig().Timeout
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != cfg.Path {
			next.ServeHTTP(w, r)
			return
		}

		status := "ok"
		upstreamStatus := ""
		httpStatus := http.StatusOK

		if cfg.UpstreamURL != "" {
			client := &http.Client{Timeout: cfg.Timeout}
			resp, err := client.Get(cfg.UpstreamURL)
			if err != nil || resp.StatusCode >= 500 {
				status = "degraded"
				upstreamStatus = "unreachable"
				httpStatus = http.StatusServiceUnavailable
			} else {
				upstreamStatus = "ok"
				resp.Body.Close()
			}
		}

		payload := healthStatus{
			Status:    status,
			Upstream:  upstreamStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(payload)
	})
}
