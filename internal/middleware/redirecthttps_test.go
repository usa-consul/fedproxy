package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func httpsOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func applyRedirectHTTPS(cfg RedirectHTTPSConfig, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	RedirectHTTPS(cfg)(http.HandlerFunc(httpsOKHandler)).ServeHTTP(w, r)
	return w
}

func TestRedirectHTTPS_Disabled_PassesThrough(t *testing.T) {
	cfg := DefaultRedirectHTTPSConfig()
	cfg.Enabled = false
	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	w := applyRedirectHTTPS(cfg, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRedirectHTTPS_PlainHTTP_Redirects(t *testing.T) {
	cfg := DefaultRedirectHTTPSConfig()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	req.Host = "example.com"
	w := applyRedirectHTTPS(cfg, req)
	if w.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "https://example.com/page" {
		t.Fatalf("unexpected Location: %s", loc)
	}
}

func TestRedirectHTTPS_XForwardedProto_PassesThrough(t *testing.T) {
	cfg := DefaultRedirectHTTPSConfig()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := applyRedirectHTTPS(cfg, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRedirectHTTPS_ExemptPath_PassesThrough(t *testing.T) {
	cfg := DefaultRedirectHTTPSConfig()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/healthz", nil)
	w := applyRedirectHTTPS(cfg, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRedirectHTTPS_CustomStatusCode(t *testing.T) {
	cfg := RedirectHTTPSConfig{
		Enabled:    true,
		StatusCode: http.StatusTemporaryRedirect,
	}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	req.Host = "example.com"
	w := applyRedirectHTTPS(cfg, req)
	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", w.Code)
	}
}

func TestRedirectHTTPS_DefaultStatusCode_WhenZero(t *testing.T) {
	cfg := RedirectHTTPSConfig{Enabled: true, StatusCode: 0}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/page", nil)
	req.Host = "example.com"
	w := applyRedirectHTTPS(cfg, req)
	if w.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", w.Code)
	}
}
