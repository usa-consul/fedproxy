package middleware

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuth_NoneMode_Passes(t *testing.T) {
	h := RequireAuth(AuthConfig{Mode: AuthModeNone}, okHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuth_SAMLMode_MissingHeader(t *testing.T) {
	h := RequireAuth(AuthConfig{Mode: AuthModeSAML}, okHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/secure", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_SAMLMode_WithHeader(t *testing.T) {
	h := RequireAuth(AuthConfig{Mode: AuthModeSAML}, okHandler())
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("X-Saml-Subject", "user@agency.gov")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuth_PIVMode_NoCert(t *testing.T) {
	h := RequireAuth(AuthConfig{Mode: AuthModePIV}, okHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/secure", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_PIVMode_WithCert(t *testing.T) {
	h := RequireAuth(AuthConfig{Mode: AuthModePIV}, okHandler())
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{{}},
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuth_ExemptPath_SkipsAuth(t *testing.T) {
	cfg := AuthConfig{
		Mode:        AuthModeSAML,
		ExemptPaths: []string{"/health"},
	}
	h := RequireAuth(cfg, okHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on exempt path, got %d", rec.Code)
	}
}
