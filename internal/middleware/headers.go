package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeadersConfig holds configuration for security header injection.
type SecurityHeadersConfig struct {
	HSTSMaxAge            int
	ContentSecurityPolicy string
	ExtraHeaders          map[string]string
}

// DefaultSecurityHeadersConfig returns a conservative default configuration.
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
		ExtraHeaders:          map[string]string{},
	}
}

// SecurityHeaders injects standard security response headers and strips
// sensitive upstream headers before forwarding the response to the client.
func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Strip inbound headers that should not be trusted from clients.
			r.Header.Del("X-Forwarded-User")
			r.Header.Del("X-Remote-User")
			r.Header.Del("X-Saml-Assertion")

			// Wrap the ResponseWriter so we can inject headers before the
			// upstream handler writes its own.
			hw := &headerWriter{ResponseWriter: w, cfg: cfg, wroteHeader: false}
			next.ServeHTTP(hw, r)
			// Ensure headers are written even if the handler never called
			// WriteHeader or Write.
			hw.ensureHeaders()
		})
	}
}

type headerWriter struct {
	http.ResponseWriter
	cfg         SecurityHeadersConfig
	wroteHeader bool
}

func (hw *headerWriter) ensureHeaders() {
	if hw.wroteHeader {
		return
	}
	hw.injectHeaders()
}

func (hw *headerWriter) injectHeaders() {
	h := hw.ResponseWriter.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("X-XSS-Protection", "1; mode=block")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	if hw.cfg.HSTSMaxAge > 0 {
		h.Set("Strict-Transport-Security",
			strings.Join([]string{
				"max-age=" + itoa(hw.cfg.HSTSMaxAge),
				"includeSubDomains",
			}, "; "))
	}
	if hw.cfg.ContentSecurityPolicy != "" {
		h.Set("Content-Security-Policy", hw.cfg.ContentSecurityPolicy)
	}
	for k, v := range hw.cfg.ExtraHeaders {
		h.Set(k, v)
	}
}

func (hw *headerWriter) WriteHeader(code int) {
	if !hw.wroteHeader {
		hw.injectHeaders()
		hw.wroteHeader = true
	}
	hw.ResponseWriter.WriteHeader(code)
}

func (hw *headerWriter) Write(b []byte) (int, error) {
	if !hw.wroteHeader {
		hw.injectHeaders()
		hw.wroteHeader = true
	}
	return hw.ResponseWriter.Write(b)
}

// itoa converts an int to its decimal string representation without
// importing strconv in this file.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 12)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
