package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestJWT_ChainedWithRequestID verifies JWT middleware cooperates with RequestID.
func TestJWT_ChainedWithRequestID(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	token := makeJWT(map[string]interface{}{"sub": "dave", "exp": float64(time.Now().Add(time.Hour).Unix())}, testSecret)

	var capturedID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	h := RequestID(JWT(cfg, inner))
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if capturedID == "" {
		t.Error("expected request ID to be set in context")
	}
}

// TestJWT_ChainedWithRateLimit verifies JWT runs after rate limiting.
func TestJWT_ChainedWithRateLimit(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	rlCfg := RateLimiterConfig{Max: 5, WindowSec: 60}
	token := makeJWT(map[string]interface{}{"sub": "rl-user", "exp": float64(time.Now().Add(time.Hour).Unix())}, testSecret)

	h := NewRateLimiter(rlCfg).RateLimit(JWT(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// TestJWT_InvalidToken_BlockedBeforeHandler ensures the downstream handler is never called.
func TestJWT_InvalidToken_BlockedBeforeHandler(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	called := false
	h := JWT(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer not.a.token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if called {
		t.Error("downstream handler should not have been called with invalid token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
