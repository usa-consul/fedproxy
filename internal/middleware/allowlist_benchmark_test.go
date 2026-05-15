package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkAllowList_ExactHit measures the hot path when the request path is
// in the allow list (exact match).
func BenchmarkAllowList_ExactHit(b *testing.B) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/api/v1/health", "/api/v1/status", "/api/v2/users"}
	h := AllowList(cfg)(http.HandlerFunc(allowOKHandler))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		h.ServeHTTP(w, r)
	}
}

// BenchmarkAllowList_ExactMiss measures the path when the request is blocked.
func BenchmarkAllowList_ExactMiss(b *testing.B) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/api/v1/health", "/api/v1/status", "/api/v2/users"}
	h := AllowList(cfg)(http.HandlerFunc(allowOKHandler))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/admin/secret", nil)
		h.ServeHTTP(w, r)
	}
}

// BenchmarkAllowList_PrefixHit measures prefix matching on a matching path.
func BenchmarkAllowList_PrefixHit(b *testing.B) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/api/", "/health"}
	cfg.PrefixMatch = true
	h := AllowList(cfg)(http.HandlerFunc(allowOKHandler))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v3/documents", nil)
		h.ServeHTTP(w, r)
	}
}
