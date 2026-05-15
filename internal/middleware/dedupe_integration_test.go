package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/fedproxy/internal/middleware"
)

// TestDedupe_ChainedWithRequestID verifies that dedupe works correctly when
// composed after RequestID so the replay header is still present.
func TestDedupe_ChainedWithRequestID(t *testing.T) {
	var counter atomic.Int32
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)
		w.WriteHeader(http.StatusCreated)
	})

	h := middleware.RequestID(
		middleware.Dedupe(middleware.DedupeConfig{
			TTL:    2 * time.Second,
			Header: "X-Idempotency-Key",
		})(inner),
	)

	send := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/orders", nil)
		req.Header.Set("X-Idempotency-Key", "order-42")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}

	first := send()
	if first.Code != http.StatusCreated {
		t.Fatalf("first: expected 201, got %d", first.Code)
	}
	second := send()
	if second.Code != http.StatusCreated {
		t.Fatalf("second: expected 201, got %d", second.Code)
	}
	if second.Header().Get("X-Dedupe-Replay") != "true" {
		t.Error("expected X-Dedupe-Replay on second response")
	}
	if counter.Load() != 1 {
		t.Fatalf("handler should be called once, got %d", counter.Load())
	}
}

// TestDedupe_ChainedWithRateLimit confirms dedupe replays bypass the rate
// limiter (the inner handler is not invoked again).
func TestDedupe_ChainedWithRateLimit(t *testing.T) {
	var counter atomic.Int32
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	rl := middleware.NewRateLimiter()
	h := middleware.RateLimit(rl, middleware.RateLimitConfig{
		RequestsPerSecond: 100,
		Burst:             100,
	})(
		middleware.Dedupe(middleware.DedupeConfig{
			TTL:    2 * time.Second,
			Header: "X-Idempotency-Key",
		})(inner),
	)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/submit", nil)
		req.Header.Set("X-Idempotency-Key", "idem-rl")
		req.RemoteAddr = "10.0.0.1:9999"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
	if counter.Load() != 1 {
		t.Fatalf("expected inner handler called once, got %d", counter.Load())
	}
}
