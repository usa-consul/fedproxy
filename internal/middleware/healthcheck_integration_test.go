package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/fedproxy/internal/middleware"
)

// TestHealthCheck_ChainedWithRequestID verifies that the health endpoint
// still gets a request-id injected when chained after RequestID middleware.
func TestHealthCheck_ChainedWithRequestID(t *testing.T) {
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	cfg := middleware.DefaultHealthCheckConfig()
	chain := middleware.RequestID(middleware.HealthCheck(cfg, base))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from health endpoint, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}
}

// TestHealthCheck_ExemptFromAuth ensures the health path can be added to
// the auth exempt list so unauthenticated probes succeed.
func TestHealthCheck_ExemptFromAuth(t *testing.T) {
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	healthCfg := middleware.DefaultHealthCheckConfig()

	// Wrap with SAML auth that exempts /healthz.
	authChain := middleware.RequireAuth(middleware.AuthConfig{
		Mode:        "saml",
		ExemptPaths: []string{"/healthz"},
	}, middleware.HealthCheck(healthCfg, base))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	authChain.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for exempt health path, got %d", rec.Code)
	}
}

// TestHealthCheck_UpstreamTimeout_Returns503 checks that a slow upstream
// causes a 503 within the configured timeout window.
func TestHealthCheck_UpstreamTimeout_Returns503(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slow.Close()

	cfg := middleware.DefaultHealthCheckConfig()
	cfg.UpstreamURL = slow.URL
	cfg.Timeout = 50 * time.Millisecond // shorter than the sleep

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := middleware.HealthCheck(cfg, base)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 on timeout, got %d", rec.Code)
	}

	var payload map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&payload)
	if payload["status"] != "degraded" {
		t.Fatalf("expected degraded status, got %v", payload["status"])
	}
}
