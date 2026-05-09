package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func compressedBodyHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	})
}

func TestCompress_NoAcceptEncoding_PassesThrough(t *testing.T) {
	mw := Compress(DefaultCompressConfig())(compressedBodyHandler(`{"ok":true}`))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("expected no gzip encoding when client does not accept it")
	}
	if rec.Body.String() != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestCompress_WithAcceptEncoding_CompressesResponse(t *testing.T) {
	body := strings.Repeat("hello world ", 200)
	mw := Compress(DefaultCompressConfig())(compressedBodyHandler(body))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected Content-Encoding: gzip")
	}

	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()
	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress body: %v", err)
	}
	if string(decompressed) != body {
		t.Fatalf("decompressed body mismatch: got %q", string(decompressed))
	}
}

func TestCompress_SetsVaryHeader(t *testing.T) {
	mw := Compress(DefaultCompressConfig())(compressedBodyHandler("data"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	mw.ServeHTTP(rec, req)

	if !strings.Contains(rec.Header().Get("Vary"), "Accept-Encoding") {
		t.Fatal("expected Vary: Accept-Encoding header")
	}
}

func TestCompress_RemovesContentLength(t *testing.T) {
	mw := Compress(DefaultCompressConfig())(compressedBodyHandler("data"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Length") != "" {
		t.Fatal("expected Content-Length to be stripped for compressed response")
	}
}

func TestDefaultCompressConfig(t *testing.T) {
	cfg := DefaultCompressConfig()
	if cfg.Level != gzip.DefaultCompression {
		t.Errorf("expected default compression level, got %d", cfg.Level)
	}
	if cfg.MinLength != 1024 {
		t.Errorf("expected MinLength 1024, got %d", cfg.MinLength)
	}
	if len(cfg.Types) == 0 {
		t.Error("expected non-empty default types list")
	}
}
