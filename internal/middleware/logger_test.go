package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/fedproxy/internal/middleware"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestRequestLogger_LogsMethod(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := middleware.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("expected log to contain method GET, got: %s", output)
	}
	if !strings.Contains(output, "/health") {
		t.Errorf("expected log to contain path /health, got: %s", output)
	}
}

func TestRequestLogger_CapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := middleware.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !strings.Contains(buf.String(), "404") {
		t.Errorf("expected log to contain status 404, got: %s", buf.String())
	}
}

func TestResponseRecorder_DefaultStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	rec := middleware.NewResponseRecorder(rr)

	if rec.StatusCode != http.StatusOK {
		t.Errorf("expected default status 200, got %d", rec.StatusCode)
	}
}

func TestResponseRecorder_TracksWrittenBytes(t *testing.T) {
	rr := httptest.NewRecorder()
	rec := middleware.NewResponseRecorder(rr)

	body := []byte("hello")
	_, _ = rec.Write(body)

	if rec.Written != int64(len(body)) {
		t.Errorf("expected %d bytes written, got %d", len(body), rec.Written)
	}
}
