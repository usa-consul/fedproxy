package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func TestSecurityHeaders_DefaultHeadersPresent(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()
	handler := SecurityHeaders(cfg)(http.HandlerFunc(echoHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for k, v := range expected {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
}

func TestSecurityHeaders_HSTSHeader(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()
	handler := SecurityHeaders(cfg)(http.HandlerFunc(echoHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	hsts := rec.Header().Get("Strict-Transport-Security")
	if hsts == "" {
		t.Fatal("expected Strict-Transport-Security header, got empty")
	}
}

func TestSecurityHeaders_StripsSensitiveRequestHeaders(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()

	var capturedReq *http.Request
	capture := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})
	handler := SecurityHeaders(cfg)(capture)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-User", "attacker")
	req.Header.Set("X-Remote-User", "attacker")
	req.Header.Set("X-Saml-Assertion", "forged")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	for _, h := range []string{"X-Forwarded-User", "X-Remote-User", "X-Saml-Assertion"} {
		if capturedReq.Header.Get(h) != "" {
			t.Errorf("expected header %s to be stripped, but it was present", h)
		}
	}
}

func TestSecurityHeaders_ExtraHeaders(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()
	cfg.ExtraHeaders = map[string]string{
		"X-Custom-Agency": "DOD",
	}
	handler := SecurityHeaders(cfg)(http.HandlerFunc(echoHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Custom-Agency"); got != "DOD" {
		t.Errorf("X-Custom-Agency = %q, want %q", got, "DOD")
	}
}

func TestSecurityHeaders_NoHSTSWhenZero(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()
	cfg.HSTSMaxAge = 0
	handler := SecurityHeaders(cfg)(http.HandlerFunc(echoHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("expected no HSTS header when max-age=0, got %q", got)
	}
}
