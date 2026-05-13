package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var healthPassthroughHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
})

func TestHealthCheck_NonHealthPath_PassesThrough(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	h := HealthCheck(cfg, healthPassthroughHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
}

func TestHealthCheck_HealthPath_Returns200(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	h := HealthCheck(cfg, healthPassthroughHandler)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHealthCheck_ResponseIsJSON(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	h := HealthCheck(cfg, healthPassthroughHandler)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("expected JSON content-type, got %q", ct)
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", payload["status"])
	}
	if _, ok := payload["timestamp"]; !ok {
		t.Fatal("expected timestamp field in response")
	}
}

func TestHealthCheck_NoCacheHeader(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	h := HealthCheck(cfg, healthPassthroughHandler)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("expected Cache-Control: no-store")
	}
}

func TestHealthCheck_UnreachableUpstream_Returns503(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	cfg.UpstreamURL = "http://127.0.0.1:19999/health" // nothing listening
	h := HealthCheck(cfg, healthPassthroughHandler)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var payload map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&payload)
	if payload["status"] != "degraded" {
		t.Fatalf("expected status degraded, got %v", payload["status"])
	}
}

func TestHealthCheck_ReachableUpstream_Returns200(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	cfg := DefaultHealthCheckConfig()
	cfg.UpstreamURL = upstream.URL
	h := HealthCheck(cfg, healthPassthroughHandler)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&payload)
	if payload["upstream"] != "ok" {
		t.Fatalf("expected upstream ok, got %v", payload["upstream"])
	}
}
