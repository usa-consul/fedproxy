package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestMetrics_ChainedWithRequestID ensures the metrics middleware cooperates
// with RequestID without double-counting or panicking.
func TestMetrics_ChainedWithRequestID(t *testing.T) {
	resetMetrics()

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	chain := RequestID(Metrics(DefaultMetricsConfig())(inner))

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := atomic.LoadUint64(&metricsState.requests); got != 1 {
		t.Fatalf("expected 1 request, got %d", got)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}
}

// TestMetrics_MultipleRequests_AccumulatesCorrectly sends several requests
// including one error and checks that all counters are accurate.
func TestMetrics_MultipleRequests_AccumulatesCorrectly(t *testing.T) {
	resetMetrics()

	callCount := 0
	handler := Metrics(DefaultMetricsConfig())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount == 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/item", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	if got := atomic.LoadUint64(&metricsState.requests); got != 5 {
		t.Fatalf("expected 5 requests, got %d", got)
	}
	if got := atomic.LoadUint64(&metricsState.errors); got != 1 {
		t.Fatalf("expected 1 error, got %d", got)
	}
}

// TestMetrics_SnapshotReflectsCounters verifies the JSON snapshot returned
// by the metrics endpoint matches the actual counters after some traffic.
func TestMetrics_SnapshotReflectsCounters(t *testing.T) {
	resetMetrics()

	handler := Metrics(DefaultMetricsConfig())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	req := httptest.NewRequest(http.MethodGet, "/_metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var snap MetricsSnapshot
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if snap.TotalRequests != 4 {
		t.Fatalf("expected TotalRequests=4, got %d", snap.TotalRequests)
	}
	if snap.Uptime == "" {
		t.Fatal("expected non-empty uptime")
	}
}
