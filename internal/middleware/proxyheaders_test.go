package middleware_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/fedproxy/internal/middleware"
)

func proxyHeaderCaptureHandler(headers *http.Header) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*headers = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})
}

func applyProxyHeaders(cfg middleware.ProxyHeadersConfig, r *http.Request) (*httptest.ResponseRecorder, http.Header) {
	var captured http.Header
	h := middleware.ProxyHeaders(cfg)(proxyHeaderCaptureHandler(&captured))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w, captured
}

func TestProxyHeaders_SetsXForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.42:54321"
	_, h := applyProxyHeaders(middleware.DefaultProxyHeadersConfig(), r)
	if got := h.Get("X-Forwarded-For"); got != "203.0.113.42" {
		t.Fatalf("X-Forwarded-For = %q, want %q", got, "203.0.113.42")
	}
}

func TestProxyHeaders_SetsXRealIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:9000"
	_, h := applyProxyHeaders(middleware.DefaultProxyHeadersConfig(), r)
	if got := h.Get("X-Real-IP"); got != "10.0.0.5" {
		t.Fatalf("X-Real-IP = %q, want %q", got, "10.0.0.5")
	}
}

func TestProxyHeaders_HTTPProto(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	_, h := applyProxyHeaders(middleware.DefaultProxyHeadersConfig(), r)
	if got := h.Get("X-Forwarded-Proto"); got != "http" {
		t.Fatalf("X-Forwarded-Proto = %q, want "http"", got)
	}
}

func TestProxyHeaders_TLSProto(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	r.TLS = &tls.ConnectionState{}
	_, h := applyProxyHeaders(middleware.DefaultProxyHeadersConfig(), r)
	if got := h.Get("X-Forwarded-Proto"); got != "https" {
		t.Fatalf("X-Forwarded-Proto = %q, want "https"", got)
	}
}

func TestProxyHeaders_ExplicitProto(t *testing.T) {
	cfg := middleware.DefaultProxyHeadersConfig()
	cfg.ForwardedProto = "https"
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	_, h := applyProxyHeaders(cfg, r)
	if got := h.Get("X-Forwarded-Proto"); got != "https" {
		t.Fatalf("X-Forwarded-Proto = %q, want "https"", got)
	}
}

func TestProxyHeaders_SetsXForwardedHost(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://example.gov/api", nil)
	r.RemoteAddr = "192.168.1.1:8080"
	_, h := applyProxyHeaders(middleware.DefaultProxyHeadersConfig(), r)
	if got := h.Get("X-Forwarded-Host"); got != "example.gov" {
		t.Fatalf("X-Forwarded-Host = %q, want "example.gov"", got)
	}
}

func TestProxyHeaders_TrustIncoming_AppendsXFF(t *testing.T) {
	cfg := middleware.DefaultProxyHeadersConfig()
	cfg.TrustIncoming = true
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.1.2.3:5000"
	r.Header.Set("X-Forwarded-For", "198.51.100.1")
	_, h := applyProxyHeaders(cfg, r)
	if got := h.Get("X-Forwarded-For"); got != "198.51.100.1, 10.1.2.3" {
		t.Fatalf("X-Forwarded-For = %q", got)
	}
}
