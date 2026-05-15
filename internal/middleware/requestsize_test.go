package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/fedproxy/internal/middleware"
)

func sizeOKHandler(w http.ResponseWriter, r *http.Request) {
	// Drain body so MaxBytesReader can do its work.
	_, _ = io.Copy(io.Discard, r.Body)
	w.WriteHeader(http.StatusOK)
}

func applyRequestSize(cfg middleware.RequestSizeConfig, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	middleware.RequestSize(cfg)(http.HandlerFunc(sizeOKHandler)).ServeHTTP(w, r)
	return w
}

func TestRequestSize_UnderAllLimits_Passes(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("hello"))
	w := applyRequestSize(cfg, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequestSize_URITooLong_Returns414(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	longPath := "/" + strings.Repeat("a", int(cfg.MaxURIBytes)+1)
	req := httptest.NewRequest(http.MethodGet, longPath, nil)
	req.RequestURI = longPath
	w := applyRequestSize(cfg, req)
	if w.Code != http.StatusRequestURITooLong {
		t.Fatalf("expected 414, got %d", w.Code)
	}
}

func TestRequestSize_HeadersTooLarge_Returns431(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Add a header that pushes total past MaxHeaderBytes.
	req.Header.Set("X-Big", strings.Repeat("x", int(cfg.MaxHeaderBytes)+1))
	w := applyRequestSize(cfg, req)
	if w.Code != http.StatusRequestHeaderFieldsTooLarge {
		t.Fatalf("expected 431, got %d", w.Code)
	}
}

func TestRequestSize_ContentLengthExceeds_Returns413(t *testing.T) {
	cfg := middleware.DefaultRequestSizeConfig()
	body := bytes.NewReader(make([]byte, 10))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = cfg.MaxBodyBytes + 1
	w := applyRequestSize(cfg, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", w.Code)
	}
}

func TestRequestSize_ZeroLimits_AllowsAll(t *testing.T) {
	cfg := middleware.RequestSizeConfig{} // all zeros → disabled
	bigURI := "/" + strings.Repeat("z", 10_000)
	req := httptest.NewRequest(http.MethodGet, bigURI, nil)
	req.RequestURI = bigURI
	w := applyRequestSize(cfg, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequestSize_CustomBodyLimit_Enforced(t *testing.T) {
	cfg := middleware.RequestSizeConfig{MaxBodyBytes: 16}
	body := strings.NewReader(strings.Repeat("b", 17))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = 17
	w := applyRequestSize(cfg, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", w.Code)
	}
}
