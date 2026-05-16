package middleware

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func reqlogOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func applyReqLog(cfg RequestLogConfig, h http.Handler) http.Handler {
	return RequestAuditLog(cfg)(h)
}

func TestRequestAuditLog_LogsBasicFields(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultRequestLogConfig()
	cfg.Logger = log.New(&buf, "", 0)

	h := applyReqLog(cfg, http.HandlerFunc(reqlogOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	out := buf.String()
	for _, want := range []string{"method", "GET", "path", "/api/data", "status", "200"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected log to contain %q, got: %s", want, out)
		}
	}
}

func TestRequestAuditLog_SkipsExemptPath(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultRequestLogConfig()
	cfg.Logger = log.New(&buf, "", 0)
	cfg.SkipPaths = []string{"/healthz"}

	h := applyReqLog(cfg, http.HandlerFunc(reqlogOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if buf.Len() != 0 {
		t.Errorf("expected no log output for skipped path, got: %s", buf.String())
	}
}

func TestRequestAuditLog_IncludesRequestIDFromHeader(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultRequestLogConfig()
	cfg.Logger = log.New(&buf, "", 0)

	h := applyReqLog(cfg, http.HandlerFunc(reqlogOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "test-req-123")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(buf.String(), "test-req-123") {
		t.Errorf("expected request_id in log, got: %s", buf.String())
	}
}

func TestRequestAuditLog_IncludesRequestIDFromContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := DefaultRequestLogConfig()
	cfg.Logger = log.New(&buf, "", 0)

	h := applyReqLog(cfg, http.HandlerFunc(reqlogOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey{}, "ctx-id-456"))
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(buf.String(), "ctx-id-456") {
		t.Errorf("expected context request_id in log, got: %s", buf.String())
	}
}

func TestRequestAuditLog_NilLogger_UsesDefault(t *testing.T) {
	cfg := RequestLogConfig{IncludeRequestID: true}
	// Should not panic when Logger is nil.
	h := applyReqLog(cfg, http.HandlerFunc(reqlogOKHandler))
	req := httptest.NewRequest(http.MethodGet, "/safe", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
}
