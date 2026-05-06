package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("something went wrong")
}

func normalHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func TestRecover_NoPanic_PassesThrough(t *testing.T) {
	mw := Recover(DefaultRecoverConfig())
	h := mw(http.HandlerFunc(normalHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRecover_Panic_Returns500(t *testing.T) {
	mw := Recover(DefaultRecoverConfig())
	h := mw(http.HandlerFunc(panicHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestRecover_Panic_LogsPanicValue(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	cfg := RecoverConfig{Logger: logger, PrintStack: false}
	mw := Recover(cfg)
	h := mw(http.HandlerFunc(panicHandler))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), "something went wrong") {
		t.Fatalf("expected panic value in log output, got: %s", buf.String())
	}
}

func TestRecover_Panic_LogsStack(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	cfg := RecoverConfig{Logger: logger, PrintStack: true}
	mw := Recover(cfg)
	h := mw(http.HandlerFunc(panicHandler))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), "stack") {
		t.Fatalf("expected stack trace in log output, got: %s", buf.String())
	}
}

func TestRecover_NilLogger_UsesDefault(t *testing.T) {
	cfg := RecoverConfig{Logger: nil, PrintStack: false}
	mw := Recover(cfg)
	h := mw(http.HandlerFunc(panicHandler))

	// Should not panic itself; just verify it returns 500.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
