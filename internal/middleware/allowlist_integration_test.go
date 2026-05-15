package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAllowList_ChainedWithRequestID verifies that a blocked request still
// receives a Request-ID header injected by the upstream RequestID middleware.
func TestAllowList_ChainedWithRequestID(t *testing.T) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/allowed"}

	chain := RequestID(AllowList(cfg)(http.HandlerFunc(allowOKHandler)))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/blocked", nil)
	chain.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if w.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}
}

// TestAllowList_ChainedWithAuth verifies that the allowlist fires before auth
// so that blocked paths are rejected without auth overhead.
func TestAllowList_ChainedWithAuth(t *testing.T) {
	allowCfg := DefaultAllowListConfig()
	allowCfg.Paths = []string{"/public"}

	authCalled := false
	sentinel := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCalled = true
		w.WriteHeader(http.StatusOK)
	})

	chain := AllowList(allowCfg)(sentinel)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/private", nil)
	chain.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if authCalled {
		t.Fatal("inner handler should not have been called for blocked path")
	}
}

// TestAllowList_MultiplePrefixes ensures several prefix rules work together.
func TestAllowList_MultiplePrefixes(t *testing.T) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/api/", "/health", "/metrics"}
	cfg.PrefixMatch = true

	cases := []struct {
		path   string
		wantOK bool
	}{
		{"/api/v1/users", true},
		{"/health", true},
		{"/metrics", true},
		{"/admin/secret", false},
		{"/internal", false},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, tc.path, nil)
		AllowList(cfg)(http.HandlerFunc(allowOKHandler)).ServeHTTP(w, r)
		gotOK := w.Code == http.StatusOK
		if gotOK != tc.wantOK {
			t.Errorf("path %s: wantOK=%v got status %d", tc.path, tc.wantOK, w.Code)
		}
	}
}
