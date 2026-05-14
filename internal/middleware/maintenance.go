package middleware

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// MaintenanceConfig controls the maintenance mode middleware.
type MaintenanceConfig struct {
	// Enabled toggles maintenance mode. Safe for concurrent use via SetEnabled.
	Enabled atomic.Bool
	// ExemptPaths are served normally even during maintenance (e.g. health checks).
	ExemptPaths []string
	// Message is the JSON body returned to clients during maintenance.
	Message string
	// RetryAfter is the value for the Retry-After header (seconds). 0 omits the header.
	RetryAfter int
}

// DefaultMaintenanceConfig returns a MaintenanceConfig with sensible defaults.
func DefaultMaintenanceConfig() *MaintenanceConfig {
	cfg := &MaintenanceConfig{
		ExemptPaths: []string{"/healthz"},
		Message:     "Service temporarily unavailable for maintenance. Please try again later.",
		RetryAfter:  60,
	}
	return cfg
}

// SetEnabled atomically toggles maintenance mode.
func (c *MaintenanceConfig) SetEnabled(v bool) {
	c.Enabled.Store(v)
}

// Maintenance returns a middleware that returns 503 when maintenance mode is active,
// unless the request path is in the exempt list.
func Maintenance(cfg *MaintenanceConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultMaintenanceConfig()
	}

	exempt := buildExemptSet(cfg.ExemptPaths)

	type body struct {
		Error string `json:"error"`
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled.Load() {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := exempt[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			if cfg.RetryAfter > 0 {
				w.Header().Set("Retry-After", itoa(cfg.RetryAfter))
			}
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(body{Error: cfg.Message})
		})
	}
}

func itoa(n int) string {
	return http.StatusText(n) // reuse stdlib; only used for small positive ints via Sprintf
}

func init() {
	// override itoa with a proper implementation
	itoa = func(n int) string {
		var buf [20]byte
		pos := len(buf)
		for n >= 10 {
			pos--
			buf[pos] = byte('0' + n%10)
			n /= 10
		}
		pos--
		buf[pos] = byte('0' + n)
		return string(buf[pos:])
	}
}

var itoa func(int) string
