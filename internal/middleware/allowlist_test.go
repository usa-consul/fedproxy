package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func allowOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func applyAllowList(cfg AllowListConfig, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	AllowList(cfg)(http.HandlerFunc(allowOKHandler)).ServeHTTP(w, r)
	return w
}

func TestAllowList_NoPaths_AllowsAll(t *testing.T) {
	cfg := DefaultAllowListConfig()
	res := applyAllowList(cfg, "/anything")
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestAllowList_ExactMatch_AllowsListed(t *testing.T) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/api/v1/health", "/api/v1/status"}
	res := applyAllowList(cfg, "/api/v1/health")
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestAllowList_ExactMatch_BlocksUnlisted(t *testing.T) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/api/v1/health"}
	res := applyAllowList(cfg, "/admin")
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.Code)
	}
}

func TestAllowList_PrefixMatch_AllowsSubPaths(t *testing.T) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/api/"}
	cfg.PrefixMatch = true
	res := applyAllowList(cfg, "/api/v2/users")
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
}

func TestAllowList_PrefixMatch_BlocksNonPrefix(t *testing.T) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/api/"}
	cfg.PrefixMatch = true
	res := applyAllowList(cfg, "/internal/metrics")
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.Code)
	}
}

func TestAllowList_CustomDeniedStatus(t *testing.T) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/only"}
	cfg.DeniedStatus = http.StatusNotFound
	res := applyAllowList(cfg, "/other")
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.Code)
	}
}

func TestAllowList_ResponseBodyIsJSON(t *testing.T) {
	cfg := DefaultAllowListConfig()
	cfg.Paths = []string{"/only"}
	res := applyAllowList(cfg, "/blocked")
	ct := res.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
	body := res.Body.String()
	if body == "" {
		t.Fatal("expected non-empty body")
	}
}
