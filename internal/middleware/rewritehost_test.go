package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// hostCaptureHandler records the Host header and X-Forwarded-Host seen by
// the upstream handler.
type hostCaptureHandler struct {
	host           string
	forwardedHost  string
}

func (h *hostCaptureHandler) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	h.host = r.Host
	h.forwardedHost = r.Header.Get("X-Forwarded-Host")
}

func applyRewriteHost(cfg RewriteHostConfig, incomingHost string) *hostCaptureHandler {
	cap := &hostCaptureHandler{}
	mw := RewriteHost(cfg)(cap)
	req := httptest.NewRequest(http.MethodGet, "http://"+incomingHost+"/", nil)
	req.Host = incomingHost
	mw.ServeHTTP(httptest.NewRecorder(), req)
	return cap
}

func TestRewriteHost_NoConfig_PassesThrough(t *testing.T) {
	cap := applyRewriteHost(DefaultRewriteHostConfig(), "original.example.com")
	if cap.host != "original.example.com" {
		t.Fatalf("expected original host, got %q", cap.host)
	}
	if cap.forwardedHost != "" {
		t.Fatalf("expected no X-Forwarded-Host, got %q", cap.forwardedHost)
	}
}

func TestRewriteHost_StaticHost_Overrides(t *testing.T) {
	cfg := RewriteHostConfig{StaticHost: "upstream.internal"}
	cap := applyRewriteHost(cfg, "public.example.com")
	if cap.host != "upstream.internal" {
		t.Fatalf("expected upstream.internal, got %q", cap.host)
	}
	if cap.forwardedHost != "public.example.com" {
		t.Fatalf("expected X-Forwarded-Host=public.example.com, got %q", cap.forwardedHost)
	}
}

func TestRewriteHost_RuleMatch_RewritesHost(t *testing.T) {
	cfg := RewriteHostConfig{
		Rules: []HostRewriteRule{
			{From: "old.example.com", To: "new.example.com"},
		},
	}
	cap := applyRewriteHost(cfg, "old.example.com")
	if cap.host != "new.example.com" {
		t.Fatalf("expected new.example.com, got %q", cap.host)
	}
	if cap.forwardedHost != "old.example.com" {
		t.Fatalf("expected X-Forwarded-Host=old.example.com, got %q", cap.forwardedHost)
	}
}

func TestRewriteHost_RuleNoMatch_PassesThrough(t *testing.T) {
	cfg := RewriteHostConfig{
		Rules: []HostRewriteRule{
			{From: "other.example.com", To: "new.example.com"},
		},
	}
	cap := applyRewriteHost(cfg, "unrelated.example.com")
	if cap.host != "unrelated.example.com" {
		t.Fatalf("expected original host, got %q", cap.host)
	}
}

func TestRewriteHost_FirstRuleWins(t *testing.T) {
	cfg := RewriteHostConfig{
		Rules: []HostRewriteRule{
			{From: "multi.example.com", To: "first.internal"},
			{From: "multi.example.com", To: "second.internal"},
		},
	}
	cap := applyRewriteHost(cfg, "multi.example.com")
	if cap.host != "first.internal" {
		t.Fatalf("expected first.internal, got %q", cap.host)
	}
}

func TestRewriteHost_CaseInsensitiveMatch(t *testing.T) {
	cfg := RewriteHostConfig{
		Rules: []HostRewriteRule{
			{From: "Case.Example.COM", To: "target.internal"},
		},
	}
	cap := applyRewriteHost(cfg, "case.example.com")
	if cap.host != "target.internal" {
		t.Fatalf("expected target.internal, got %q", cap.host)
	}
}
