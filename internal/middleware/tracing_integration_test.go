package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/fedproxy/internal/middleware"
)

// TestTracing_ChainedWithRequestID verifies that Tracing and RequestID
// middleware coexist without header conflicts.
func TestTracing_ChainedWithRequestID(t *testing.T) {
	h := middleware.RequestID(
		middleware.Tracing(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-Request-Id") == "" {
		t.Error("expected X-Request-Id header")
	}
	if rr.Header().Get("Traceparent") == "" {
		t.Error("expected Traceparent header")
	}
}

// TestTracing_UniqueIDsPerRequest ensures each request receives a distinct trace.
func TestTracing_UniqueIDsPerRequest(t *testing.T) {
	h := middleware.Tracing(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	collect := func() string {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		return rr.Header().Get("Traceparent")
	}

	a, b := collect(), collect()
	if a == b {
		t.Errorf("expected unique traceparents per request, both were %q", a)
	}
}

// TestTracing_InvalidTraceparent_FallsBackToNewTrace checks graceful handling
// of a malformed incoming header.
func TestTracing_InvalidTraceparent_FallsBackToNewTrace(t *testing.T) {
	h := middleware.Tracing(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Traceparent", "not-a-valid-header")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	got := rr.Header().Get("Traceparent")
	if !strings.HasPrefix(got, "00-") {
		t.Errorf("expected valid traceparent fallback, got %q", got)
	}
	// The generated trace should not echo back the invalid input.
	if strings.Contains(got, "not-a-valid") {
		t.Errorf("unexpected invalid content in traceparent: %q", got)
	}
	fmt.Println("fallback traceparent:", got)
}
