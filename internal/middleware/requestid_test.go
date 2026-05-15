package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesIDWhenAbsent(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if capturedID == "" {
		t.Fatal("expected a request ID on context, got empty string")
	}
	if len(capturedID) != 32 {
		t.Fatalf("expected 32-char hex ID, got %q (len %d)", capturedID, len(capturedID))
	}
}

func TestRequestID_ReusesIncomingHeader(t *testing.T) {
	const existingID = "abc123"

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, existingID)

	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if capturedID != existingID {
		t.Fatalf("expected ID %q, got %q", existingID, capturedID)
	}
}

func TestRequestID_SetsResponseHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get(RequestIDHeader); got == "" {
		t.Fatal("expected X-Request-ID response header to be set")
	}
}

func TestRequestID_ForwardsIDToUpstream(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	var upstreamID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamID = r.Header.Get(RequestIDHeader)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if upstreamID == "" {
		t.Fatal("expected X-Request-ID to be forwarded in request headers")
	}
	if upstreamID != rr.Header().Get(RequestIDHeader) {
		t.Fatalf("upstream header %q does not match response header %q", upstreamID, rr.Header().Get(RequestIDHeader))
	}
}

func TestFromContext_EmptyWhenNotSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if id := FromContext(req.Context()); id != "" {
		t.Fatalf("expected empty string, got %q", id)
	}
}

func TestRequestID_UniqueIDsPerRequest(t *testing.T) {
	ids := make(map[string]struct{}, 10)

	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids[FromContext(r.Context())] = struct{}{}
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 10; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(rr, req)
	}

	if len(ids) != 10 {
		t.Fatalf("expected 10 unique request IDs, got %d", len(ids))
	}
}
