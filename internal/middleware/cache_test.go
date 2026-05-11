package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func cacheOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func cacheErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestCache_DefaultConfig_SetsMaxAge(t *testing.T) {
	cfg := DefaultCacheConfig()
	h := Cache(cfg)(http.HandlerFunc(cacheOKHandler))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	h.ServeHTTP(rr, req)

	cc := rr.Header().Get("Cache-Control")
	if cc == "" {
		t.Fatal("expected Cache-Control header to be set")
	}
	if cc == "no-store" {
		t.Errorf("expected public max-age directive, got: %s", cc)
	}
}

func TestCache_PrivatePath_SetsNoStore(t *testing.T) {
	cfg := DefaultCacheConfig()
	h := Cache(cfg)(http.HandlerFunc(cacheOKHandler))

	for _, path := range []string{"/auth/login", "/saml/acs", "/piv/verify"} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		h.ServeHTTP(rr, req)

		cc := rr.Header().Get("Cache-Control")
		if cc != "no-store" {
			t.Errorf("path %s: expected no-store, got %q", path, cc)
		}
	}
}

func TestCache_NonCacheableStatus_SetsNoStore(t *testing.T) {
	cfg := DefaultCacheConfig()
	h := Cache(cfg)(http.HandlerFunc(cacheErrorHandler))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	h.ServeHTTP(rr, req)

	cc := rr.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("expected no-store for 500 response, got %q", cc)
	}
}

func TestCache_ZeroMaxAge_SetsNoStore(t *testing.T) {
	cfg := CacheConfig{
		MaxAge:            0,
		CacheableStatuses: []int{http.StatusOK},
	}
	h := Cache(cfg)(http.HandlerFunc(cacheOKHandler))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	h.ServeHTTP(rr, req)

	cc := rr.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("expected no-store when MaxAge is zero, got %q", cc)
	}
}

func TestCache_CustomMaxAge(t *testing.T) {
	cfg := CacheConfig{
		MaxAge:            5 * time.Minute,
		CacheableStatuses: []int{http.StatusOK},
	}
	h := Cache(cfg)(http.HandlerFunc(cacheOKHandler))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/public/asset", nil)
	h.ServeHTTP(rr, req)

	cc := rr.Header().Get("Cache-Control")
	expected := "public, max-age=300"
	if cc != expected {
		t.Errorf("expected %q, got %q", expected, cc)
	}
}

func TestIsPrivatePath(t *testing.T) {
	prefixes := []string{"/auth", "/saml"}
	cases := []struct {
		path    string
		want    bool
	}{
		{"/auth/login", true},
		{"/saml/metadata", true},
		{"/public", false},
		{"/", false},
	}
	for _, tc := range cases {
		got := isPrivatePath(tc.path, prefixes)
		if got != tc.want {
			t.Errorf("isPrivatePath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
