package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/fedproxy/internal/middleware"
)

func TestRequestID_ChainedWithLogger(t *testing.T) {
	var capturedID string

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = middleware.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	chain := middleware.RequestID(middleware.RequestLogger(inner))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	chain.ServeHTTP(rw, req)

	if capturedID == "" {
		t.Fatal("expected request ID to be set in context")
	}
	if rw.Header().Get("X-Request-Id") != capturedID {
		t.Errorf("response header X-Request-Id = %q, want %q", rw.Header().Get("X-Request-Id"), capturedID)
	}
}

func TestRequestID_ChainedWithAuth_IDPresentInBothModes(t *testing.T) {
	var seenID string

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenID = middleware.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	cfg := middleware.DefaultAuthConfig()
	cfg.Mode = "none"

	chain := middleware.RequestID(middleware.RequireAuth(cfg, inner))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rw := httptest.NewRecorder()
	chain.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	if seenID == "" {
		t.Error("expected request ID to be visible inside auth middleware")
	}
}

func TestRequestID_UniquePerConcurrentRequest(t *testing.T) {
	const n = 20
	ids := make(chan string, n)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids <- middleware.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	chain := middleware.RequestID(inner)

	for i := 0; i < n; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rw := httptest.NewRecorder()
			chain.ServeHTTP(rw, req)
		}()
	}

	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := <-ids
		if _, dup := seen[id]; dup {
			t.Errorf("duplicate request ID: %q", id)
		}
		seen[id] = struct{}{}
	}
}
