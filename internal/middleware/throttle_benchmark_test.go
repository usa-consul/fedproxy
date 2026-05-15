package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkThrottle_LowContention(b *testing.B) {
	cfg := ThrottleConfig{
		MaxConcurrent: 1000,
		QueueTimeout:  time.Second,
	}
	h := Throttle(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		}
	})
}

func BenchmarkThrottle_HighContention(b *testing.B) {
	cfg := ThrottleConfig{
		MaxConcurrent: 4,
		QueueTimeout:  500 * time.Millisecond,
	}
	h := Throttle(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
		}
	})
}
