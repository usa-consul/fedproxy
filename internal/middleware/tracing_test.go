package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func tracingEchoHandler(t *testing.T, gotInfo *TraceInfo) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, _ := TraceFromContext(r.Context())
		*gotInfo = info
		w.WriteHeader(http.StatusOK)
	})
}

func TestTracing_GeneratesTraceWhenAbsent(t *testing.T) {
	var info TraceInfo
	h := Tracing(tracingEchoHandler(t, &info))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if info.TraceID == "" {
		t.Fatal("expected TraceID to be set")
	}
	if info.SpanID == "" {
		t.Fatal("expected SpanID to be set")
	}
	if info.Parent != "" {
		t.Errorf("expected no parent for root trace, got %q", info.Parent)
	}
}

func TestTracing_PropagatesIncomingTraceparent(t *testing.T) {
	var info TraceInfo
	h := Tracing(tracingEchoHandler(t, &info))

	traceID := strings.Repeat("a", 32)
	parentSpan := strings.Repeat("b", 16)
	incoming := fmt.Sprintf("00-%s-%s-01", traceID, parentSpan)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Traceparent", incoming)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if info.TraceID != traceID {
		t.Errorf("expected TraceID %q, got %q", traceID, info.TraceID)
	}
	if info.Parent != parentSpan {
		t.Errorf("expected Parent %q, got %q", parentSpan, info.Parent)
	}
	if info.SpanID == parentSpan {
		t.Error("child SpanID must differ from parent SpanID")
	}
}

func TestTracing_SetsResponseHeader(t *testing.T) {
	var info TraceInfo
	h := Tracing(tracingEchoHandler(t, &info))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	got := rr.Header().Get("Traceparent")
	if got == "" {
		t.Fatal("expected Traceparent response header")
	}
	if !strings.HasPrefix(got, "00-") {
		t.Errorf("unexpected traceparent format: %q", got)
	}
}

func TestTracing_ForwardsTraceparentUpstream(t *testing.T) {
	var upstreamHeader string
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHeader = r.Header.Get("Traceparent")
		w.WriteHeader(http.StatusOK)
	})
	h := Tracing(upstream)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if upstreamHeader == "" {
		t.Fatal("Traceparent header not forwarded to upstream")
	}
}

func TestFromContext_EmptyWhenTracingNotSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := TraceFromContext(req.Context())
	if ok {
		t.Error("expected no TraceInfo in context without Tracing middleware")
	}
}
