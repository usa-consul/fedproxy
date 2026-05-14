package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/fedproxy/internal/middleware"
)

// TestGeoBlock_ChainedWithRequestID verifies GeoBlock works inside a typical
// middleware chain and that blocked requests never reach the inner handler.
func TestGeoBlock_ChainedWithRequestID(t *testing.T) {
	reached := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})

	geoCfg := middleware.GeoBlockConfig{
		CountryCodes: []string{"KP"},
		Block:        true,
		Lookup:       fixedLookup("KP"),
	}

	chain := middleware.RequestID(middleware.GeoBlock(geoCfg)(inner))

	r := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 from GeoBlock, got %d", w.Code)
	}
	if reached {
		t.Error("inner handler must not be reached when geo-blocked")
	}
	if w.Header().Get("X-Request-Id") == "" {
		t.Error("X-Request-Id should still be set by RequestID middleware")
	}
}

// TestGeoBlock_AllowMode_ChainedWithAuth ensures allowlist mode integrates
// correctly and that an allowed country proceeds to subsequent middleware.
func TestGeoBlock_AllowMode_ChainedWithAuth(t *testing.T) {
	reached := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})

	geoCfg := middleware.GeoBlockConfig{
		CountryCodes: []string{"US"},
		Block:        false,
		Lookup:       fixedLookup("US"),
	}

	chain := middleware.GeoBlock(geoCfg)(inner)
	r := httptest.NewRequest(http.MethodGet, "/secure", nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for allowed country, got %d", w.Code)
	}
	if !reached {
		t.Error("inner handler should have been reached for allowed country")
	}
}

// TestGeoBlock_CaseInsensitiveCodes verifies that lowercase country codes in
// config are normalised and still match correctly.
func TestGeoBlock_CaseInsensitiveCodes(t *testing.T) {
	cfg := middleware.GeoBlockConfig{
		CountryCodes: []string{"ru", "cn"},
		Block:        true,
		Lookup:       fixedLookup("RU"),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	middleware.GeoBlock(cfg)(http.HandlerFunc(geoOKHandler)).ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for RU (config has lowercase 'ru'), got %d", w.Code)
	}
}
