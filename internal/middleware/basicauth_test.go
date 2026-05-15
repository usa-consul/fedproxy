package middleware

import (
	"encoding/base64"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func basicOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func applyBasicAuth(cfg BasicAuthConfig, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	cfg.Logger = log.New(os.Stderr, "", 0)
	BasicAuth(cfg, http.HandlerFunc(basicOKHandler)).ServeHTTP(w, r)
	return w
}

func basicHeader(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func TestBasicAuth_MissingHeader_Returns401(t *testing.T) {
	cfg := DefaultBasicAuthConfig()
	cfg.Credentials = map[string]string{"alice": "secret"}
	r := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := applyBasicAuth(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Error("expected WWW-Authenticate header")
	}
}

func TestBasicAuth_ValidCredentials_Passes(t *testing.T) {
	cfg := DefaultBasicAuthConfig()
	cfg.Credentials = map[string]string{"alice": "secret"}
	r := httptest.NewRequest(http.MethodGet, "/api", nil)
	r.Header.Set("Authorization", basicHeader("alice", "secret"))
	w := applyBasicAuth(cfg, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBasicAuth_WrongPassword_Returns401(t *testing.T) {
	cfg := DefaultBasicAuthConfig()
	cfg.Credentials = map[string]string{"alice": "secret"}
	r := httptest.NewRequest(http.MethodGet, "/api", nil)
	r.Header.Set("Authorization", basicHeader("alice", "wrong"))
	w := applyBasicAuth(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestBasicAuth_UnknownUser_Returns401(t *testing.T) {
	cfg := DefaultBasicAuthConfig()
	cfg.Credentials = map[string]string{"alice": "secret"}
	r := httptest.NewRequest(http.MethodGet, "/api", nil)
	r.Header.Set("Authorization", basicHeader("bob", "secret"))
	w := applyBasicAuth(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestBasicAuth_ExemptPath_Passes(t *testing.T) {
	cfg := DefaultBasicAuthConfig()
	cfg.Credentials = map[string]string{"alice": "secret"}
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	// no Authorization header
	w := applyBasicAuth(cfg, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on exempt path, got %d", w.Code)
	}
}

func TestBasicAuth_EmptyCredentials_AlwaysDenies(t *testing.T) {
	cfg := DefaultBasicAuthConfig()
	// no credentials configured
	r := httptest.NewRequest(http.MethodGet, "/api", nil)
	r.Header.Set("Authorization", basicHeader("alice", "secret"))
	w := applyBasicAuth(cfg, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with empty credential store, got %d", w.Code)
	}
}
