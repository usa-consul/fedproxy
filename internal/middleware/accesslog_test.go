package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func writeBodyHandler(body string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

func TestAccessLog_LogsRequestLine(t *testing.T) {
	var logged string
	cfg := DefaultAccessLogConfig()
	cfg.LogFunc = func(line string) { logged = line }

	h := AccessLog(cfg)(writeBodyHandler("ok", http.StatusOK))
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	for _, want := range []string{"GET", "/api/data", "200"} {
		if !strings.Contains(logged, want) {
			t.Errorf("expected log to contain %q, got: %s", want, logged)
		}
	}
}

func TestAccessLog_SkipsConfiguredPaths(t *testing.T) {
	var logged string
	cfg := DefaultAccessLogConfig()
	cfg.SkipPaths = []string{"/healthz"}
	cfg.LogFunc = func(line string) { logged = line }

	h := AccessLog(cfg)(writeBodyHandler("ok", http.StatusOK))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if logged != "" {
		t.Errorf("expected no log for skipped path, got: %s", logged)
	}
}

func TestAccessLog_LogsNon200Status(t *testing.T) {
	var logged string
	cfg := DefaultAccessLogConfig()
	cfg.LogFunc = func(line string) { logged = line }

	h := AccessLog(cfg)(writeBodyHandler("not found", http.StatusNotFound))
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(logged, "404") {
		t.Errorf("expected log to contain status 404, got: %s", logged)
	}
}

func TestAccessLog_IncludesRequestAndTraceIDs(t *testing.T) {
	var logged string
	cfg := DefaultAccessLogConfig()
	cfg.LogFunc = func(line string) { logged = line }

	// Chain RequestID + Tracing so context values are populated.
	h := RequestID(AccessLog(cfg)(Tracing(writeBodyHandler("ok", http.StatusOK))))
	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	for _, want := range []string{"req=", "trace="} {
		if !strings.Contains(logged, want) {
			t.Errorf("expected log to contain %q, got: %s", want, logged)
		}
	}
}

func TestAccessLog_DefaultLogFuncUsedWhenNil(t *testing.T) {
	cfg := AccessLogConfig{} // LogFunc intentionally nil
	h := AccessLog(cfg)(writeBodyHandler("ok", http.StatusOK))
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	// Should not panic when LogFunc is nil — default is applied internally.
	h.ServeHTTP(httptest.NewRecorder(), req)
}
