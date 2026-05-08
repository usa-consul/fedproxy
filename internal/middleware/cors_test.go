package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_NoOriginHeader_PassesThrough(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"https://example.gov"}
	h := CORS(cfg)(okHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS header when Origin is absent")
	}
}

func TestCORS_DisallowedOrigin_Returns403(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"https://allowed.gov"}
	h := CORS(cfg)(okHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestCORS_AllowedOrigin_SetsHeaders(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"https://portal.gov"}
	h := CORS(cfg)(okHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://portal.gov")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://portal.gov" {
		t.Errorf("unexpected Allow-Origin: %q", got)
	}
}

func TestCORS_WildcardOrigin_Allowed(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"*"}
	h := CORS(cfg)(okHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://anything.example")
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected wildcard, got %q", got)
	}
}

func TestCORS_PreflightOptions_Returns204(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"https://portal.gov"}
	h := CORS(cfg)(okHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/resource", nil)
	req.Header.Set("Origin", "https://portal.gov")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods to be set on preflight")
	}
}

func TestCORS_Credentials_SetsHeader(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"https://portal.gov"}
	cfg.AllowCredentials = true
	h := CORS(cfg)(okHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://portal.gov")
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected credentials header, got %q", got)
	}
}
