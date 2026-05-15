package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestThrottle_ChainedWithRequestID verifies that a throttled request still
// carries the X-Request-ID header injected by the RequestID middleware.
func TestThrottle_ChainedWithRequestID(t *testing.T) {
	cfg := DefaultThrottleConfig()

	chain := RequestID(Throttle(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := FromContext(r.Context())
		if id == "" {
			t.Error("expected request ID in context, got empty string")
		}
		w.WriteHeader(http.StatusOK)
	})))

	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/data", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID response header")
	}
}

// TestThrottle_ChainedWithRateLimit verifies that throttle and rate-limit
// coexist: a request passing the rate limiter is still throttled correctly.
func TestThrottle_ChainedWithRateLimit(t *testing.T) {
	rl := NewRateLimiter(10, time.Second)
	tCfg := ThrottleConfig{
		MaxConcurrent: 5,
		QueueTimeout:  100 * time.Millisecond,
		StatusCode:    http.StatusServiceUnavailable,
	}

	chain := RateLimit(rl, Throttle(tCfg, http.HandlerFunc(throttleOKHandler)))

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rr := httptest.NewRecorder()
			chain.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
			if rr.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rr.Code)
			}
		}()
	}
	wg.Wait()
}

// TestThrottle_503CustomStatus verifies a custom status code is returned on
// queue exhaustion when the caller overrides StatusCode.
func TestThrottle_503CustomStatus(t *testing.T) {
	blocked := make(chan struct{})
	release := make(chan struct{})

	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(blocked)
		<-release
		w.WriteHeader(http.StatusOK)
	})

	cfg := ThrottleConfig{
		MaxConcurrent: 1,
		QueueTimeout:  30 * time.Millisecond,
		StatusCode:    http.StatusTooManyRequests,
	}
	h := Throttle(cfg, slow)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()

	<-blocked
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	close(release)
	wg.Wait()
}
