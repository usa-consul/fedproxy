package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// resetMetrics zeroes the global counters between tests.
func resetMetrics() {
	atomic.StoreUint64(&metricsState.requests, 0)
	atomic.StoreUint64(&metricsState.errors, 0)
	atomic.StoreUint64(&metricsState.bytes, 0)
	atomic.StoreUint64(&metricsState.latencySum, 0)
}

func TestMetrics_NonMetricsPath_PassesThrough(t *testing.T) {
	resetMetrics()
	handler := Metrics(DefaultMetricsConfig())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := atomic.LoadUint64(&metricsState.requests); got != 1 {
		t.Fatalf("expected 1 request counted, got %d", got)
	}
}

func TestMetrics_MetricsPath_ReturnsJSON(t *testing.T) {
	resetMetrics()
	handler := Metrics(DefaultMetricsConfig())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/_metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
	var snap MetricsSnapshot
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode metrics JSON: %v", err)
	}
}

func TestMetrics_CountsErrors(t *testing.T) {
	resetMetrics()
	handler := Metrics(DefaultMetricsConfig())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if got := atomic.LoadUint64(&metricsState.errors); got != 1 {
		t.Fatalf("expected 1 error counted, got %d", got)
	}
}

func TestMetrics_NoCacheHeader(t *testing.T) {
	resetMetrics()
	handler := Metrics(DefaultMetricsConfig())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/_metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("expected Cache-Control: no-store, got %q", cc)
	}
}

func TestMetrics_DefaultConfig(t *testing.T) {
	cfg := DefaultMetricsConfig()
	if cfg.Path != "/_metrics" {
		t.Fatalf("unexpected default path: %s", cfg.Path)
	}
}

func TestMetrics_CustomPath(t *testing.T) {
	resetMetrics()
	cfg := MetricsConfig{Path: "/internal/stats"}
	hit := false
	handler := Metrics(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/internal/stats", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if hit {
		t.Fatal("upstream should not have been called for metrics path")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
