package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/fedproxy/internal/middleware"
)

func tagCaptureHandler(t *testing.T, got *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = middleware.TagFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
}

func applyRequestTag(cfg middleware.RequestTagConfig, next http.Handler) http.Handler {
	return middleware.RequestTag(cfg)(next)
}

func TestRequestTag_NoHeader_NoTagInContext(t *testing.T) {
	var captured string
	h := applyRequestTag(middleware.DefaultRequestTagConfig(), tagCaptureHandler(t, &captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if captured != "" {
		t.Fatalf("expected empty tag, got %q", captured)
	}
}

func TestRequestTag_WithHeader_StoresInContext(t *testing.T) {
	var captured string
	h := applyRequestTag(middleware.DefaultRequestTagConfig(), tagCaptureHandler(t, &captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tag", "batch-job")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if captured != "batch-job" {
		t.Fatalf("expected tag %q, got %q", "batch-job", captured)
	}
}

func TestRequestTag_TruncatesLongValue(t *testing.T) {
	var captured string
	cfg := middleware.DefaultRequestTagConfig()
	cfg.MaxLength = 8
	h := applyRequestTag(cfg, tagCaptureHandler(t, &captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tag", "this-is-a-very-long-tag")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if len(captured) != 8 {
		t.Fatalf("expected tag length 8, got %d (%q)", len(captured), captured)
	}
}

func TestRequestTag_AllowList_BlocksUnknown(t *testing.T) {
	var captured string
	cfg := middleware.DefaultRequestTagConfig()
	cfg.Allowed = []string{"alpha", "beta"}
	h := applyRequestTag(cfg, tagCaptureHandler(t, &captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tag", "gamma")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if captured != "" {
		t.Fatalf("expected tag to be cleared, got %q", captured)
	}
}

func TestRequestTag_AllowList_PermitsKnown(t *testing.T) {
	var captured string
	cfg := middleware.DefaultRequestTagConfig()
	cfg.Allowed = []string{"alpha", "beta"}
	h := applyRequestTag(cfg, tagCaptureHandler(t, &captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tag", "ALPHA") // case-insensitive
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if captured != "ALPHA" {
		t.Fatalf("expected tag %q, got %q", "ALPHA", captured)
	}
}

func TestRequestTag_SetsResponseHeader(t *testing.T) {
	var captured string
	h := applyRequestTag(middleware.DefaultRequestTagConfig(), tagCaptureHandler(t, &captured))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tag", "echo-me")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Tag"); got != "echo-me" {
		t.Fatalf("expected response header %q, got %q", "echo-me", got)
	}
}
