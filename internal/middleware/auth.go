package middleware

import (
	"net/http"
	"strings"
)

// AuthMode defines the authentication strategy enforced by the proxy.
type AuthMode string

const (
	AuthModeNone AuthMode = "none"
	AuthModeSAML AuthMode = "saml"
	AuthModePIV  AuthMode = "piv"
)

// AuthConfig holds configuration for the auth middleware.
type AuthConfig struct {
	Mode           AuthMode
	ExemptPaths    []string
	SAMLMetadata   string
}

// RequireAuth returns an HTTP middleware that enforces the configured auth mode.
// For now, SAML and PIV modes validate the presence of expected headers/certs;
// full IdP integration is wired in later phases.
func RequireAuth(cfg AuthConfig, next http.Handler) http.Handler {
	exempt := buildExemptSet(cfg.ExemptPaths)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if exempt[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		switch cfg.Mode {
		case AuthModeSAML:
			if !hasSAMLAssertion(r) {
				http.Error(w, "SAML authentication required", http.StatusUnauthorized)
				return
			}
		case AuthModePIV:
			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				http.Error(w, "PIV certificate required", http.StatusUnauthorized)
				return
			}
		default:
			// AuthModeNone — pass through
		}

		next.ServeHTTP(w, r)
	})
}

// hasSAMLAssertion checks for a minimal SAML assertion signal on the request.
// A real implementation would validate a session cookie or signed assertion.
func hasSAMLAssertion(r *http.Request) bool {
	v := r.Header.Get("X-Saml-Subject")
	return strings.TrimSpace(v) != ""
}

func buildExemptSet(paths []string) map[string]bool {
	m := make(map[string]bool, len(paths))
	for _, p := range paths {
		m[p] = true
	}
	return m
}
