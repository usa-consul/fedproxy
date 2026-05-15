package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/fedproxy/internal/middleware"
)

func dedupeCountHandler(counter *atomic.Int32) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"ok":true}`) //nolint:errcheck
	})
}

func applyDedupe(cfg middleware.DedupeConfig, h http.Handler) http.Handler {
	return middleware.Dedupe(cfg)(h)
}

func TestDedupe_NoKey_PassesThrough(t *testing.T) {
	var counter atomic.Int32
	h := applyDedupe(middleware.DefaultDedupeConfig(), dedupeCountHandler(&counter))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
	if counter.Load() != 3 {
		t.Fatalf("expected handler called 3 times, got %d", counter.Load())
	}
}

func TestDedupe_SameKey_ReplayedFromCache(t *testing.T) {
	var counter atomic.Int32
	cfg := middleware.DedupeConfig{TTL: 2 * time.Second, Header: "X-Idempotency-Key"}
	h := applyDedupe(cfg, dedupeCountHandler(&counter))

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodPost, "/pay", nil)
		req.Header.Set("X-Idempotency-Key", "key-abc")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	}
	if counter.Load() != 1 {
		t.Fatalf("expected handler called once, got %d", counter.Load())
	}
}

func TestDedupe_ReplayedResponse_SetsHeader(t *testing.T) {
	var counter atomic.Int32
	cfg := middleware.DedupeConfig{TTL: 2 * time.Second, Header: "X-Idempotency-Key"}
	h := applyDedupe(cfg, dedupeCountHandler(&counter))

	send := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-Idempotency-Key", "key-replay")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}

	first := send()
	if first.Header().Get("X-Dedupe-Replay") != "" {
		t.Error("first response should not have replay header")
	}
	second := send()
	if second.Header().Get("X-Dedupe-Replay") != "true" {
		t.Error("second response should have X-Dedupe-Replay: true")
	}
}

func TestDedupe_ExpiredTTL_CallsHandlerAgain(t *testing.T) {
	var counter atomic.Int32
	cfg := middleware.DedupeConfig{TTL: 20 * time.Millisecond, Header: "X-Idempotency-Key"}
	h := applyDedupe(cfg, dedupeCountHandler(&counter))

	send := func() {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-Idempotency-Key", "key-ttl")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}

	send()
	time.Sleep(40 * time.Millisecond)
	send()

	if counter.Load() != 2 {
		t.Fatalf("expected 2 handler calls after TTL expiry, got %d", counter.Load())
	}
}

func TestDedupe_DifferentKeys_BothForwarded(t *testing.T) {
	var counter atomic.Int32
	cfg := middleware.DedupeConfig{TTL: 2 * time.Second, Header: "X-Idempotency-Key"}
	h := applyDedupe(cfg, dedupeCountHandler(&counter))

	for _, key := range []string{"key-1", "key-2", "key-3"} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-Idempotency-Key", key)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
	if counter.Load() != 3 {
		t.Fatalf("expected 3 handler calls for 3 distinct keys, got %d", counter.Load())
	}
}
