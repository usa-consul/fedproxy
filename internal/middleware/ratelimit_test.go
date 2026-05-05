package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllow_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !rl.Allow("127.0.0.1") {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")
	if rl.Allow("10.0.0.1") {
		t.Fatal("expected deny after limit exceeded")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Millisecond)
	rl.Allow("192.168.1.1")
	time.Sleep(20 * time.Millisecond)
	if !rl.Allow("192.168.1.1") {
		t.Fatal("expected allow after window reset")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	rl.Allow("1.1.1.1")
	if !rl.Allow("2.2.2.2") {
		t.Fatal("expected allow for different key")
	}
}

func TestRateLimit_Middleware_Passes(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	mw := RateLimit(rl)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRateLimit_Middleware_Blocks(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	mw := RateLimit(rl)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.3:5678"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req) // first: allowed
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req) // second: blocked
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	if got := clientIP(req); got != "203.0.113.5" {
		t.Fatalf("expected 203.0.113.5, got %s", got)
	}
}
