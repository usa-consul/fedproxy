package middleware_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/fedproxy/internal/middleware"
)

func geoOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func fixedLookup(country string) func(string) (string, error) {
	return func(_ string) (string, error) { return country, nil }
}

func errorLookup(_ string) (string, error) {
	return "", fmt.Errorf("lookup failed")
}

func applyGeoBlock(cfg middleware.GeoBlockConfig, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	middleware.GeoBlock(cfg)(http.HandlerFunc(geoOKHandler)).ServeHTTP(w, r)
	return w
}

func TestGeoBlock_NoCodes_AllowsAll(t *testing.T) {
	cfg := middleware.DefaultGeoBlockConfig()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyGeoBlock(cfg, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGeoBlock_BlockMode_BlocksListedCountry(t *testing.T) {
	cfg := middleware.GeoBlockConfig{
		CountryCodes: []string{"CN", "RU"},
		Block:        true,
		Lookup:       fixedLookup("CN"),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyGeoBlock(cfg, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	var body map[string]string
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["country"] != "CN" {
		t.Errorf("expected country CN in response, got %q", body["country"])
	}
}

func TestGeoBlock_BlockMode_AllowsUnlistedCountry(t *testing.T) {
	cfg := middleware.GeoBlockConfig{
		CountryCodes: []string{"CN"},
		Block:        true,
		Lookup:       fixedLookup("US"),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyGeoBlock(cfg, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGeoBlock_AllowMode_BlocksUnlistedCountry(t *testing.T) {
	cfg := middleware.GeoBlockConfig{
		CountryCodes: []string{"US"},
		Block:        false,
		Lookup:       fixedLookup("DE"),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyGeoBlock(cfg, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestGeoBlock_AllowMode_AllowsListedCountry(t *testing.T) {
	cfg := middleware.GeoBlockConfig{
		CountryCodes: []string{"US"},
		Block:        false,
		Lookup:       fixedLookup("US"),
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyGeoBlock(cfg, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGeoBlock_LookupError_FailsOpen(t *testing.T) {
	cfg := middleware.GeoBlockConfig{
		CountryCodes: []string{"CN"},
		Block:        true,
		Lookup:       errorLookup,
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := applyGeoBlock(cfg, r)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on lookup error (fail-open), got %d", w.Code)
	}
}

func TestGeoBlock_XForwardedFor_UsesFirstIP(t *testing.T) {
	seen := ""
	cfg := middleware.GeoBlockConfig{
		CountryCodes: []string{},
		Block:        true,
		Lookup: func(ip string) (string, error) {
			seen = ip
			return "US", nil
		},
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	applyGeoBlock(cfg, r)
	if seen != "1.2.3.4" {
		t.Errorf("expected first XFF IP 1.2.3.4, got %q", seen)
	}
}
