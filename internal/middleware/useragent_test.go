package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func uaOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func applyUserAgent(cfg UserAgentConfig, ua string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}
	rec := httptest.NewRecorder()
	UserAgent(cfg, http.HandlerFunc(uaOKHandler)).ServeHTTP(rec, req)
	return rec
}

func TestUserAgent_NoRules_AllowsAll(t *testing.T) {
	cfg := DefaultUserAgentConfig()
	rec := applyUserAgent(cfg, "Mozilla/5.0")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUserAgent_BlockedSubstring_Returns403(t *testing.T) {
	cfg := UserAgentConfig{BlockedAgents: []string{"curl"}}
	rec := applyUserAgent(cfg, "curl/7.68.0")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestUserAgent_BlockedCaseInsensitive(t *testing.T) {
	cfg := UserAgentConfig{BlockedAgents: []string{"badbot"}}
	rec := applyUserAgent(cfg, "BADBOT/1.0")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestUserAgent_AllowedAgent_Passes(t *testing.T) {
	cfg := UserAgentConfig{BlockedAgents: []string{"curl", "wget"}}
	rec := applyUserAgent(cfg, "Mozilla/5.0 (compatible)")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUserAgent_RequireNonEmpty_EmptyReturns400(t *testing.T) {
	cfg := UserAgentConfig{RequireNonEmpty: true}
	rec := applyUserAgent(cfg, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUserAgent_RequireNonEmpty_PresentPasses(t *testing.T) {
	cfg := UserAgentConfig{RequireNonEmpty: true}
	rec := applyUserAgent(cfg, "MyClient/2.0")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUserAgent_MultipleBlocked_MatchesFirst(t *testing.T) {
	cfg := UserAgentConfig{BlockedAgents: []string{"scanner", "sqlmap"}}
	rec := applyUserAgent(cfg, "sqlmap/1.0")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
