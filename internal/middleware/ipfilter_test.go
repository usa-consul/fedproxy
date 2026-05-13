package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func ipOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func applyIPFilter(t *testing.T, cfg IPFilterConfig, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()
	mw, err := IPFilter(cfg)
	if err != nil {
		t.Fatalf("IPFilter error: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(ipOKHandler)).ServeHTTP(rec, req)
	return rec
}

func TestIPFilter_NoRules_AllowsAll(t *testing.T) {
	rec := applyIPFilter(t, DefaultIPFilterConfig(), "203.0.113.5:1234")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestIPFilter_BlockCIDR_BlocksMatchingIP(t *testing.T) {
	cfg := IPFilterConfig{BlockCIDRs: []string{"203.0.113.0/24"}}
	rec := applyIPFilter(t, cfg, "203.0.113.42:9000")
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestIPFilter_BlockCIDR_AllowsNonMatchingIP(t *testing.T) {
	cfg := IPFilterConfig{BlockCIDRs: []string{"203.0.113.0/24"}}
	rec := applyIPFilter(t, cfg, "198.51.100.1:9000")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestIPFilter_AllowCIDR_AllowsMatchingIP(t *testing.T) {
	cfg := IPFilterConfig{AllowCIDRs: []string{"10.0.0.0/8"}}
	rec := applyIPFilter(t, cfg, "10.1.2.3:5000")
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestIPFilter_AllowCIDR_BlocksNonMatchingIP(t *testing.T) {
	cfg := IPFilterConfig{AllowCIDRs: []string{"10.0.0.0/8"}}
	rec := applyIPFilter(t, cfg, "172.16.0.1:5000")
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestIPFilter_BlockTakesPrecedenceOverAllow(t *testing.T) {
	cfg := IPFilterConfig{
		AllowCIDRs: []string{"10.0.0.0/8"},
		BlockCIDRs: []string{"10.0.0.0/24"},
	}
	// IP is in both allow and block — block wins
	rec := applyIPFilter(t, cfg, "10.0.0.5:1234")
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestIPFilter_InvalidCIDR_ReturnsError(t *testing.T) {
	cfg := IPFilterConfig{AllowCIDRs: []string{"not-a-cidr"}}
	_, err := IPFilter(cfg)
	if err == nil {
		t.Error("expected error for invalid CIDR, got nil")
	}
}

func TestIPFilter_TrustProxy_UsesXForwardedFor(t *testing.T) {
	cfg := IPFilterConfig{
		BlockCIDRs: []string{"203.0.113.0/24"},
		TrustProxy: true,
	}
	mw, err := IPFilter(cfg)
	if err != nil {
		t.Fatalf("IPFilter error: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9000"
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(ipOKHandler)).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 via XFF, got %d", rec.Code)
	}
}
