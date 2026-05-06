package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func failHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadGateway)
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected circuit closed, got: %v", err)
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 3, OpenDuration: 10 * time.Second}
	cb := NewCircuitBreaker(cfg)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if err := cb.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got: %v", err)
	}
}

func TestCircuitBreaker_ResetsOnSuccess(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 2, OpenDuration: 10 * time.Second}
	cb := NewCircuitBreaker(cfg)
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected circuit closed after success reset, got: %v", err)
	}
}

func TestCircuitBreaker_HalfOpenAfterDuration(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 1, OpenDuration: 10 * time.Millisecond}
	cb := NewCircuitBreaker(cfg)
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected half-open to allow request, got: %v", err)
	}
}

func TestCircuitBreakerMiddleware_BlocksWhenOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 1, OpenDuration: 10 * time.Second}
	cb := NewCircuitBreaker(cfg)
	cb.RecordFailure()

	h := CircuitBreakerMiddleware(cb, http.HandlerFunc(successHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rw.Code)
	}
}

func TestCircuitBreakerMiddleware_RecordsFailure(t *testing.T) {
	cfg := CircuitBreakerConfig{MaxFailures: 2, OpenDuration: 10 * time.Second}
	cb := NewCircuitBreaker(cfg)

	h := CircuitBreakerMiddleware(cb, http.HandlerFunc(failHandler))
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
	}

	if err := cb.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected circuit open after failures, got: %v", err)
	}
}
