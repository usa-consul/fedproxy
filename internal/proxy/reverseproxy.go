package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/fedproxy/internal/config"
)

// Handler wraps a reverse proxy with fedproxy-specific middleware.
type Handler struct {
	cfg    *config.Config
	proxy  *httputil.ReverseProxy
	server *http.Server
}

// New creates a new proxy Handler from the provided config.
func New(cfg *config.Config) (*Handler, error) {
	upstream, err := url.Parse(cfg.Upstream)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL %q: %w", cfg.Upstream, err)
	}

	rp := httputil.NewSingleHostReverseProxy(upstream)
	rp.ErrorHandler = defaultErrorHandler

	h := &Handler{
		cfg:   cfg,
		proxy: rp,
	}

	h.server = &http.Server{
		Addr:         cfg.Addr,
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return h, nil
}

// ServeHTTP implements http.Handler, forwarding requests to the upstream.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Inject a header so the upstream can identify fedproxy.
	r.Header.Set("X-Forwarded-By", "fedproxy")
	h.proxy.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server.
func (h *Handler) ListenAndServe() error {
	return h.server.ListenAndServe()
}

// defaultErrorHandler writes a minimal error response when the upstream is unreachable.
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, fmt.Sprintf("bad gateway: %v", err), http.StatusBadGateway)
}
