package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// slowHandler sleeps for d before responding.
func slowHandler(d time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case <-r.Context().Done():
			// context cancelled — do nothing so the middleware can respond
		}
	})
}

func TestRequestTimeout_FastHandler(t *testing.T) {
	mw := RequestTimeout(TimeoutConfig{Timeout: 500 * time.Millisecond})
	h := mw(slowHandler(10 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestTimeout_SlowHandler_Returns504(t *testing.T) {
	mw := RequestTimeout(TimeoutConfig{Timeout: 50 * time.Millisecond, Message: "timed out"})
	h := mw(slowHandler(300 * time.Millisecond))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "timed out" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestRequestTimeout_DefaultConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected 30s default, got %v", cfg.Timeout)
	}
	if cfg.Message == "" {
		t.Fatal("expected non-empty default message")
	}
}

func TestRequestTimeout_ZeroTimeout_UsesDefault(t *testing.T) {
	// A zero timeout should fall back to the default (30s), so a fast handler
	// should complete without triggering a 504.
	mw := RequestTimeout(TimeoutConfig{Timeout: 0})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
