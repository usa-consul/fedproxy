package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testSecret = "super-secret-key"

func makeJWT(claims map[string]interface{}, secret string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(header + "." + payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return header + "." + payload + "." + sig
}

func jwtOKHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, claims.Subject)
}

func TestJWT_MissingToken_Returns401(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	h := JWT(cfg, http.HandlerFunc(jwtOKHandler))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWT_ValidToken_PassesThrough(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	token := makeJWT(map[string]interface{}{"sub": "alice", "exp": float64(time.Now().Add(time.Hour).Unix())}, testSecret)
	h := JWT(cfg, http.HandlerFunc(jwtOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "alice") {
		t.Errorf("expected subject alice in body")
	}
}

func TestJWT_ExpiredToken_Returns401(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	cfg.ClockSkewSec = 0
	token := makeJWT(map[string]interface{}{"sub": "bob", "exp": float64(time.Now().Add(-time.Hour).Unix())}, testSecret)
	h := JWT(cfg, http.HandlerFunc(jwtOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWT_WrongSecret_Returns401(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	token := makeJWT(map[string]interface{}{"sub": "eve", "exp": float64(time.Now().Add(time.Hour).Unix())}, "wrong-secret")
	h := JWT(cfg, http.HandlerFunc(jwtOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWT_ExemptPath_SkipsValidation(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	h := JWT(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on exempt path, got %d", rec.Code)
	}
}

func TestJWT_ClaimsStoredInContext(t *testing.T) {
	cfg := DefaultJWTConfig(testSecret)
	token := makeJWT(map[string]interface{}{"sub": "carol", "iss": "fedproxy", "exp": float64(time.Now().Add(time.Hour).Unix())}, testSecret)
	h := JWT(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok || claims.Issuer != "fedproxy" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
