package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
)

// BasicAuthConfig holds configuration for HTTP Basic Authentication middleware.
type BasicAuthConfig struct {
	// Credentials maps username to password (plaintext for simplicity; use env vars).
	Credentials map[string]string
	// Realm is the authentication realm presented in the WWW-Authenticate header.
	Realm string
	// ExemptPaths are paths that bypass authentication (e.g. health, metrics).
	ExemptPaths []string
	// Logger is used for auth failure logging.
	Logger *log.Logger
}

// DefaultBasicAuthConfig returns a BasicAuthConfig with sensible defaults.
func DefaultBasicAuthConfig() BasicAuthConfig {
	return BasicAuthConfig{
		Credentials: map[string]string{},
		Realm:       "fedproxy",
		ExemptPaths: []string{"/healthz", "/__metrics"},
	}
}

// BasicAuth returns middleware that enforces HTTP Basic Authentication.
func BasicAuth(cfg BasicAuthConfig, next http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	exempt := buildExemptSet(cfg.ExemptPaths)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := exempt[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		username, password, ok := parseBasicAuth(r)
		if !ok {
			cfg.Logger.Printf("basicauth: missing credentials from %s", r.RemoteAddr)
			challenge(w, cfg.Realm)
			return
		}

		expected, exists := cfg.Credentials[username]
		if !exists || subtle.ConstantTimeCompare([]byte(password), []byte(expected)) != 1 {
			cfg.Logger.Printf("basicauth: invalid credentials for user %q from %s", username, r.RemoteAddr)
			challenge(w, cfg.Realm)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseBasicAuth(r *http.Request) (username, password string, ok bool) {
	hdr := r.Header.Get("Authorization")
	if !strings.HasPrefix(hdr, "Basic ") {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(hdr, "Basic "))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func challenge(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
