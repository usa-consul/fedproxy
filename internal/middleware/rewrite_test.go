package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// pathCaptureHandler records the URL path it receives.
func pathCaptureHandler(captured *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
}

func TestPathRewrite_NoRules_PassesThrough(t *testing.T) {
	var got string
	handler := PathRewrite(DefaultRewriteConfig())(pathCaptureHandler(&got))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/api/v1/resource" {
		t.Errorf("expected /api/v1/resource, got %s", got)
	}
}

func TestPathRewrite_StripPrefix(t *testing.T) {
	var got string
	cfg := RewriteConfig{
		Rules: []RewriteRule{{StripPrefix: "/api/v1", AddPrefix: ""}},
	}
	handler := PathRewrite(cfg)(pathCaptureHandler(&got))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/resource" {
		t.Errorf("expected /resource, got %s", got)
	}
}

func TestPathRewrite_StripAndAdd(t *testing.T) {
	var got string
	cfg := RewriteConfig{
		Rules: []RewriteRule{{StripPrefix: "/api/v1", AddPrefix: "/v2"}},
	}
	handler := PathRewrite(cfg)(pathCaptureHandler(&got))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/v2/users" {
		t.Errorf("expected /v2/users, got %s", got)
	}
}

func TestPathRewrite_FirstMatchWins(t *testing.T) {
	var got string
	cfg := RewriteConfig{
		Rules: []RewriteRule{
			{StripPrefix: "/api", AddPrefix: "/first"},
			{StripPrefix: "/api", AddPrefix: "/second"},
		},
	}
	handler := PathRewrite(cfg)(pathCaptureHandler(&got))
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/first/health" {
		t.Errorf("expected /first/health, got %s", got)
	}
}

func TestPathRewrite_NoMatch_PassesThrough(t *testing.T) {
	var got string
	cfg := RewriteConfig{
		Rules: []RewriteRule{{StripPrefix: "/internal", AddPrefix: "/svc"}},
	}
	handler := PathRewrite(cfg)(pathCaptureHandler(&got))
	req := httptest.NewRequest(http.MethodGet, "/public/page", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got != "/public/page" {
		t.Errorf("expected /public/page, got %s", got)
	}
}

func TestPathRewrite_OriginalRequestUnmodified(t *testing.T) {
	originalPath := "/api/v1/data"
	cfg := RewriteConfig{
		Rules: []RewriteRule{{StripPrefix: "/api/v1", AddPrefix: "/backend"}},
	}
	var got string
	handler := PathRewrite(cfg)(pathCaptureHandler(&got))
	req := httptest.NewRequest(http.MethodGet, originalPath, nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if req.URL.Path != originalPath {
		t.Errorf("original request mutated: got %s, want %s", req.URL.Path, originalPath)
	}
}
