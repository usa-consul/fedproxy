package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fedproxy/internal/config"
)

func minimalConfig(upstream string) *config.Config {
	return &config.Config{
		Addr:     ":8080",
		Upstream: upstream,
	}
}

func TestNew_ValidUpstream(t *testing.T) {
	h, err := New(minimalConfig("http://localhost:9090"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNew_InvalidUpstream(t *testing.T) {
	_, err := New(minimalConfig("://bad-url"))
	if err == nil {
		t.Fatal("expected error for invalid upstream URL, got nil")
	}
}

func TestServeHTTP_ForwardsRequest(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Forwarded-By"); got != "fedproxy" {
			t.Errorf("expected X-Forwarded-By=fedproxy, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	h, err := New(minimalConfig(backend.URL))
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestServeHTTP_BadGatewayOnUnreachableUpstream(t *testing.T) {
	h, err := New(minimalConfig("http://127.0.0.1:1")) // nothing listening
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}
