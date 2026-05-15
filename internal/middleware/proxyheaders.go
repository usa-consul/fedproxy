package middleware

import (
	"net/http"
	"strings"
)

// ProxyHeadersConfig controls which forwarding headers are injected.
type ProxyHeadersConfig struct {
	// TrustIncoming allows already-set X-Forwarded-* headers to pass through
	// rather than being overwritten. Disable in public-facing deployments.
	TrustIncoming bool

	// SetXRealIP injects the X-Real-IP header with the direct remote address.
	SetXRealIP bool

	// ForwardedProto is the scheme to advertise in X-Forwarded-Proto.
	// If empty it is inferred from the incoming request.
	ForwardedProto string
}

// DefaultProxyHeadersConfig returns a safe default configuration.
func DefaultProxyHeadersConfig() ProxyHeadersConfig {
	return ProxyHeadersConfig{
		TrustIncoming: false,
		SetXRealIP:    true,
		ForwardedProto: "",
	}
}

// ProxyHeaders injects standard forwarding headers (X-Forwarded-For,
// X-Forwarded-Proto, X-Real-IP) before passing the request upstream.
// It should be placed early in the middleware chain, after RequestID.
func ProxyHeaders(cfg ProxyHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Derive the client IP from RemoteAddr (strip port).
			remoteIP := r.RemoteAddr
			if idx := strings.LastIndex(remoteIP, ":"); idx != -1 {
				remoteIP = remoteIP[:idx]
			}
			remoteIP = strings.Trim(remoteIP, "[]")

			// X-Forwarded-For: append or set.
			if cfg.TrustIncoming && r.Header.Get("X-Forwarded-For") != "" {
				existing := r.Header.Get("X-Forwarded-For")
				r.Header.Set("X-Forwarded-For", existing+", "+remoteIP)
			} else {
				r.Header.Set("X-Forwarded-For", remoteIP)
			}

			// X-Real-IP: direct remote address.
			if cfg.SetXRealIP {
				r.Header.Set("X-Real-IP", remoteIP)
			}

			// X-Forwarded-Proto: prefer explicit config, then TLS, then http.
			proto := cfg.ForwardedProto
			if proto == "" {
				if !cfg.TrustIncoming || r.Header.Get("X-Forwarded-Proto") == "" {
					if r.TLS != nil {
						proto = "https"
					} else {
						proto = "http"
					}
				}
			}
			if proto != "" {
				r.Header.Set("X-Forwarded-Proto", proto)
			}

			// X-Forwarded-Host: preserve the original Host header.
			if r.Host != "" && (!cfg.TrustIncoming || r.Header.Get("X-Forwarded-Host") == "") {
				r.Header.Set("X-Forwarded-Host", r.Host)
			}

			next.ServeHTTP(w, r)
		})
	}
}
