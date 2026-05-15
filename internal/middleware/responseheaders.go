package middleware

import (
	"net/http"
	"strings"
)

// ResponseHeadersConfig controls which headers are added to, removed from,
// or overwritten on every upstream response before it reaches the client.
type ResponseHeadersConfig struct {
	// Add appends headers without replacing existing values.
	Add map[string]string
	// Set overwrites (or creates) headers unconditionally.
	Set map[string]string
	// Remove strips headers from the upstream response.
	Remove []string
}

// DefaultResponseHeadersConfig returns an empty, no-op configuration.
func DefaultResponseHeadersConfig() ResponseHeadersConfig {
	return ResponseHeadersConfig{
		Add:    make(map[string]string),
		Set:    make(map[string]string),
		Remove: nil,
	}
}

// ResponseHeaders returns middleware that mutates upstream response headers
// according to cfg before writing them to the client.
func ResponseHeaders(cfg ResponseHeadersConfig) func(http.Handler) http.Handler {
	// Normalise the Remove list once at construction time.
	removeSet := make(map[string]struct{}, len(cfg.Remove))
	for _, h := range cfg.Remove {
		removeSet[strings.ToLower(h)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseHeadersRecorder{
				ResponseWriter: w,
				cfg:            cfg,
				removeSet:      removeSet,
				headersSent:    false,
			}
			next.ServeHTTP(rw, r)
		})
	}
}

// responseHeadersRecorder wraps ResponseWriter and applies mutations lazily,
// just before the status code is written so that upstream headers are visible.
type responseHeadersRecorder struct {
	http.ResponseWriter
	cfg         ResponseHeadersConfig
	removeSet   map[string]struct{}
	headersSent bool
}

func (r *responseHeadersRecorder) WriteHeader(code int) {
	r.applyMutations()
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseHeadersRecorder) Write(b []byte) (int, error) {
	if !r.headersSent {
		r.applyMutations()
	}
	return r.ResponseWriter.Write(b)
}

func (r *responseHeadersRecorder) applyMutations() {
	if r.headersSent {
		return
	}
	r.headersSent = true

	h := r.ResponseWriter.Header()

	for _, name := range r.cfg.Remove {
		h.Del(name)
	}
	for k, v := range r.cfg.Set {
		h.Set(k, v)
	}
	for k, v := range r.cfg.Add {
		h.Add(k, v)
	}
}
