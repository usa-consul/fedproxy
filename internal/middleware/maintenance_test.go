package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func maintenanceOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func applyMaintenance(cfg *MaintenanceConfig, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	Maintenance(cfg)(http.HandlerFunc(maintenanceOKHandler)).ServeHTTP(rr, req)
	return rr
}

func TestMaintenance_Disabled_PassesThrough(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.SetEnabled(false)

	rr := applyMaintenance(cfg, httptest.NewRequest(http.MethodGet, "/api/data", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestMaintenance_Enabled_Returns503(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.SetEnabled(true)

	rr := applyMaintenance(cfg, httptest.NewRequest(http.MethodGet, "/api/data", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestMaintenance_Enabled_ReturnsJSON(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.SetEnabled(true)

	rr := applyMaintenance(cfg, httptest.NewRequest(http.MethodGet, "/", nil))

	var payload map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := payload["error"]; !ok {
		t.Error("expected 'error' key in JSON response")
	}
}

func TestMaintenance_ExemptPath_PassesThrough(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.SetEnabled(true)

	rr := applyMaintenance(cfg, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected exempt path to return 200, got %d", rr.Code)
	}
}

func TestMaintenance_RetryAfterHeader(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.SetEnabled(true)
	cfg.RetryAfter = 120

	rr := applyMaintenance(cfg, httptest.NewRequest(http.MethodGet, "/api", nil))
	if got := rr.Header().Get("Retry-After"); got != "120" {
		t.Errorf("expected Retry-After: 120, got %q", got)
	}
}

func TestMaintenance_ZeroRetryAfter_OmitsHeader(t *testing.T) {
	cfg := DefaultMaintenanceConfig()
	cfg.SetEnabled(true)
	cfg.RetryAfter = 0

	rr := applyMaintenance(cfg, httptest.NewRequest(http.MethodGet, "/api", nil))
	if got := rr.Header().Get("Retry-After"); got != "" {
		t.Errorf("expected no Retry-After header, got %q", got)
	}
}

func TestMaintenance_NilConfig_UsesDefaults(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	// nil config should not panic and should use defaults (maintenance off)
	Maintenance(nil)(http.HandlerFunc(maintenanceOKHandler)).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with nil config (disabled by default), got %d", rr.Code)
	}
}
