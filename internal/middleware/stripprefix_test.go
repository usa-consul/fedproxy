package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/fedproxy/internal/middleware"
)

func stripPathHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Seen-Path", r.URL.Path)
		w.Header().Set("X-Seen-Prefix", r.Header.Get("X-Forwarded-Prefix"))
		w.WriteHeader(http.StatusOK)
	})
}

func applyStripPrefix(cfg middleware.StripPrefixConfig, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	middleware.StripPrefix(cfg, stripPathHandler(nil)).ServeHTTP(w, r)
	return w
}

func TestStripPrefix_NoMatch_PassesThrough(t *testing.T) {
	cfg := middleware.StripPrefixConfig{Prefixes: []string{"/api"}}
	req := httptest.NewRequest(http.MethodGet, "/other/resource", nil)
	w := httptest.NewRecorder()
	middleware.StripPrefix(cfg, stripPathHandler(t)).ServeHTTP(w, req)
	if got := w.Header().Get("X-Seen-Path"); got != "/other/resource" {
		t.Errorf("expected path /other/resource, got %s", got)
	}
	if got := w.Header().Get("X-Seen-Prefix"); got != "" {
		t.Errorf("expected no forwarded prefix, got %s", got)
	}
}

func TestStripPrefix_MatchingPrefix_IsStripped(t *testing.T) {
	cfg := middleware.StripPrefixConfig{Prefixes: []string{"/api"}}
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	middleware.StripPrefix(cfg, stripPathHandler(t)).ServeHTTP(w, req)
	if got := w.Header().Get("X-Seen-Path"); got != "/users" {
		t.Errorf("expected path /users, got %s", got)
	}
	if got := w.Header().Get("X-Seen-Prefix"); got != "/api" {
		t.Errorf("expected forwarded prefix /api, got %s", got)
	}
}

func TestStripPrefix_StripsToRoot_NormalisesToSlash(t *testing.T) {
	cfg := middleware.StripPrefixConfig{Prefixes: []string{"/api"}}
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()
	middleware.StripPrefix(cfg, stripPathHandler(t)).ServeHTTP(w, req)
	if got := w.Header().Get("X-Seen-Path"); got != "/" {
		t.Errorf("expected normalised path /, got %s", got)
	}
}

func TestStripPrefix_FirstMatchWins(t *testing.T) {
	cfg := middleware.StripPrefixConfig{Prefixes: []string{"/api", "/api/v1"}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	w := httptest.NewRecorder()
	middleware.StripPrefix(cfg, stripPathHandler(t)).ServeHTTP(w, req)
	// "/api" matches first so result is "/v1/items", not "/items".
	if got := w.Header().Get("X-Seen-Path"); got != "/v1/items" {
		t.Errorf("expected /v1/items, got %s", got)
	}
}

func TestStripPrefix_NoPrefixes_PassesThrough(t *testing.T) {
	cfg := middleware.DefaultStripPrefixConfig()
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	w := httptest.NewRecorder()
	middleware.StripPrefix(cfg, stripPathHandler(t)).ServeHTTP(w, req)
	if got := w.Header().Get("X-Seen-Path"); got != "/anything" {
		t.Errorf("expected /anything, got %s", got)
	}
}

func TestStripPrefix_EmptyPrefixEntry_IsSkipped(t *testing.T) {
	cfg := middleware.StripPrefixConfig{Prefixes: []string{"", "/svc"}}
	req := httptest.NewRequest(http.MethodGet, "/svc/health", nil)
	w := httptest.NewRecorder()
	middleware.StripPrefix(cfg, stripPathHandler(t)).ServeHTTP(w, req)
	if got := w.Header().Get("X-Seen-Path"); got != "/health" {
		t.Errorf("expected /health, got %s", got)
	}
}
