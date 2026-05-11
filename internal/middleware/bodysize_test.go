package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func readBodyHandler(w http.ResponseWriter, r *http.Request) {
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func TestLimitBody_UnderLimit_Passes(t *testing.T) {
	cfg := BodySizeConfig{MaxBytes: 100}
	h := LimitBody(cfg)(http.HandlerFunc(readBodyHandler))

	body := strings.NewReader("small body")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestLimitBody_ContentLengthExceeds_Returns413(t *testing.T) {
	cfg := BodySizeConfig{MaxBytes: 10}
	h := LimitBody(cfg)(http.HandlerFunc(readBodyHandler))

	body := strings.NewReader("this body is definitely over ten bytes")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = int64(body.Len())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestLimitBody_StreamingExceedsLimit_Returns500OrHandled(t *testing.T) {
	cfg := BodySizeConfig{MaxBytes: 5}
	h := LimitBody(cfg)(http.HandlerFunc(readBodyHandler))

	// Send without Content-Length so the early check is skipped.
	body := bytes.NewReader([]byte("more than five bytes here"))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// MaxBytesReader causes ReadAll to fail; our handler returns 500.
	if rec.Code == http.StatusOK {
		t.Fatal("expected non-200 response when streaming body exceeds limit")
	}
}

func TestLimitBody_ZeroLimit_PassesThrough(t *testing.T) {
	cfg := BodySizeConfig{MaxBytes: 0}
	h := LimitBody(cfg)(http.HandlerFunc(readBodyHandler))

	big := strings.NewReader(strings.Repeat("x", 10_000))
	req := httptest.NewRequest(http.MethodPost, "/", big)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with no limit, got %d", rec.Code)
	}
}

func TestDefaultBodySizeConfig(t *testing.T) {
	cfg := DefaultBodySizeConfig()
	if cfg.MaxBytes != 1<<20 {
		t.Fatalf("expected default 1 MB, got %d", cfg.MaxBytes)
	}
}
