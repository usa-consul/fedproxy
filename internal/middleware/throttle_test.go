package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func throttleOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func applyThrottle(cfg ThrottleConfig, h http.HandlerFunc) http.Handler {
	return Throttle(cfg, http.HandlerFunc(h))
}

func TestThrottle_UnderLimit_Passes(t *testing.T) {
	cfg := DefaultThrottleConfig()
	h := applyThrottle(cfg, throttleOKHandler)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestThrottle_ExceedsLimit_Returns503(t *testing.T) {
	cfg := ThrottleConfig{
		MaxConcurrent: 1,
		QueueTimeout:  50 * time.Millisecond,
		StatusCode:    http.StatusServiceUnavailable,
	}

	blocked := make(chan struct{})
	release := make(chan struct{})

	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(blocked)
		<-release
		w.WriteHeader(http.StatusOK)
	})

	h := Throttle(cfg, slow)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()

	<-blocked // first request holds the slot

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}

	close(release)
	wg.Wait()
}

func TestThrottle_SlotReleasedAfterRequest(t *testing.T) {
	cfg := ThrottleConfig{
		MaxConcurrent: 1,
		QueueTimeout:  200 * time.Millisecond,
	}
	h := applyThrottle(cfg, throttleOKHandler)

	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, rr.Code)
		}
	}
}

func TestThrottle_ConcurrentWithinLimit(t *testing.T) {
	const limit = 5
	cfg := ThrottleConfig{
		MaxConcurrent: limit,
		QueueTimeout:  500 * time.Millisecond,
	}

	var active int64
	h := Throttle(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt64(&active, 1)
		if cur > limit {
			t.Errorf("active=%d exceeded limit=%d", cur, limit)
		}
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt64(&active, -1)
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}
	wg.Wait()
}

func TestThrottle_DefaultConfig_ZeroValues(t *testing.T) {
	// zero-value config should fall back to defaults
	h := Throttle(ThrottleConfig{}, http.HandlerFunc(throttleOKHandler))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
