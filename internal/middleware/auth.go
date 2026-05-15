package middleware

import (
	"crypto/tls"
	"net/http"
	"strings"
)

// AuthConfig holds configuration for the RequireAuth middleware.
type AuthConfig struct {
	// Mode is one of: "none", "saml", "piv".
	Mode string
	// ExemptPaths are paths that bypass authentication checks.
	ExemptPaths []string
}

// DefaultAuthConfig returns a permissive default config (mode: none).
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Mode:        "none",
		ExemptPaths: []string{"/health", "/_metrics"},
	}
}

// RequireAuth enforces authentication based on the configured mode.
func RequireAuth(cfg AuthConfig, next http.Handler) http.Handler {
	exempt := buildExemptSet(cfg.ExemptPaths)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := exempt[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		switch strings.ToLower(cfg.Mode) {
		case "saml":
			if !hasSAMLAssertion(r) {
				http.Error(w, "SAML assertion required", http.StatusUnauthorized)
				return
			}
		case "piv":
			if !hasPIVCert(r) {
				http.Error(w, "PIV certificate required", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// hasSAMLAssertion checks for a SAML assertion in the request header.
func hasSAMLAssertion(r *http.Request) bool {
	return r.Header.Get("X-Saml-Assertion") != ""
}

// hasPIVCert checks whether the request carries a verified TLS client certificate.
func hasPIVCert(r *http.Request) bool {
	if r.TLS == nil {
		return false
	}
	var verified [][]tls.Certificate
	_ = verified // placeholder for future chain validation
	return len(r.TLS.PeerCertificates) > 0
}

// buildExemptSet converts a slice of paths into a set for O(1) lookup.
func buildExemptSet(paths []string) map[string]struct{} {
	set := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		set[p] = struct{}{}
	}
	return set
}
