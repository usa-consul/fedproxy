package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func rhOKHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Upstream", "present")
	w.Header().Set("Server", "internal-server/1.0")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func applyResponseHeaders(cfg ResponseHeadersConfig, h http.HandlerFunc) *httptest.ResponseRecorder {
	mw := ResponseHeaders(cfg)(http.HandlerFunc(h))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rec, req)
	return rec
}

func TestResponseHeaders_DefaultConfig_NoMutation(t *testing.T) {
	cfg := DefaultResponseHeadersConfig()
	rec := applyResponseHeaders(cfg, rhOKHandler)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Upstream") != "present" {
		t.Error("expected upstream header to be preserved")
	}
}

func TestResponseHeaders_SetOverwritesHeader(t *testing.T) {
	cfg := DefaultResponseHeadersConfig()
	cfg.Set["Server"] = "fedproxy/1.0"

	rec := applyResponseHeaders(cfg, rhOKHandler)

	if got := rec.Header().Get("Server"); got != "fedproxy/1.0" {
		t.Errorf("expected Server=fedproxy/1.0, got %q", got)
	}
}

func TestResponseHeaders_AddAppendsHeader(t *testing.T) {
	cfg := DefaultResponseHeadersConfig()
	cfg.Add["X-Custom"] = "injected"

	rec := applyResponseHeaders(cfg, rhOKHandler)

	if got := rec.Header().Get("X-Custom"); got != "injected" {
		t.Errorf("expected X-Custom=injected, got %q", got)
	}
	// Upstream header must still be present.
	if rec.Header().Get("X-Upstream") != "present" {
		t.Error("expected upstream header to survive Add")
	}
}

func TestResponseHeaders_RemoveStripsHeader(t *testing.T) {
	cfg := DefaultResponseHeadersConfig()
	cfg.Remove = []string{"Server"}

	rec := applyResponseHeaders(cfg, rhOKHandler)

	if got := rec.Header().Get("Server"); got != "" {
		t.Errorf("expected Server to be removed, got %q", got)
	}
}

func TestResponseHeaders_CombinedOperations(t *testing.T) {
	cfg := ResponseHeadersConfig{
		Add:    map[string]string{"X-Proxy": "fedproxy"},
		Set:    map[string]string{"Server": "fedproxy/1.0"},
		Remove: []string{"X-Upstream"},
	}

	rec := applyResponseHeaders(cfg, rhOKHandler)

	if rec.Header().Get("X-Upstream") != "" {
		t.Error("X-Upstream should have been removed")
	}
	if rec.Header().Get("Server") != "fedproxy/1.0" {
		t.Error("Server should have been overwritten")
	}
	if rec.Header().Get("X-Proxy") != "fedproxy" {
		t.Error("X-Proxy should have been added")
	}
}

func TestResponseHeaders_WriteWithoutWriteHeader(t *testing.T) {
	cfg := DefaultResponseHeadersConfig()
	cfg.Set["X-Injected"] = "yes"

	// Handler writes body without explicit WriteHeader.
	h := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("body"))
	}

	rec := applyResponseHeaders(cfg, h)

	if rec.Header().Get("X-Injected") != "yes" {
		t.Error("expected X-Injected to be set even without explicit WriteHeader")
	}
}
