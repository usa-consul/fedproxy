package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "fedproxy-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	_ = f.Close()
	return f.Name()
}

func TestLoad_ValidMinimal(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("addr: got %q", cfg.Addr)
	}
}

func TestLoad_MissingAddr(t *testing.T) {
	path := writeTempConfig(t, "upstream: http://localhost:9090\n")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing addr")
	}
}

func TestLoad_MissingUpstream(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\n")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing upstream")
	}
}

func TestLoad_SAMLEnabledMissingMetadata(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\nauth:\n  mode: saml\n")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for saml without metadata_url")
	}
}

func TestLoad_InvalidAuthMode(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\nauth:\n  mode: oauth\n")
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for invalid auth mode")
	}
}

func TestLoad_TimeoutParsed(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\ntimeout:\n  request: 45s\n  message: request timed out\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timeout.Request != 45*time.Second {
		t.Errorf("timeout.request: got %v, want 45s", cfg.Timeout.Request)
	}
	if cfg.Timeout.Message != "request timed out" {
		t.Errorf("timeout.message: got %q", cfg.Timeout.Message)
	}
}

func TestLoad_RateLimitParsed(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\nrate_limit:\n  requests: 100\n  window: 1m\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RateLimit.Requests != 100 {
		t.Errorf("rate_limit.requests: got %d", cfg.RateLimit.Requests)
	}
	if cfg.RateLimit.Window != time.Minute {
		t.Errorf("rate_limit.window: got %v", cfg.RateLimit.Window)
	}
}
