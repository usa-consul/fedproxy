package config

import (
	"os"
	"testing"
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
		t.Errorf("expected :8080, got %q", cfg.Addr)
	}
}

func TestLoad_MissingAddr(t *testing.T) {
	path := writeTempConfig(t, "upstream: http://localhost:9090\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing addr")
	}
}

func TestLoad_MissingUpstream(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing upstream")
	}
}

func TestLoad_SAMLEnabledMissingMetadata(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\nauth:\n  mode: saml\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for saml mode without metadata")
	}
}

func TestLoad_RateLimitInvalidRequests(t *testing.T) {
	content := "addr: :8080\nupstream: http://localhost:9090\nrate_limit:\n  enabled: true\n  requests: 0\n  window: 1s\n"
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for rate_limit.requests == 0")
	}
}

func TestLoad_RateLimitInvalidWindow(t *testing.T) {
	content := "addr: :8080\nupstream: http://localhost:9090\nrate_limit:\n  enabled: true\n  requests: 10\n  window: 0s\n"
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for rate_limit.window == 0")
	}
}

func TestLoad_CORSEnabledNoOrigins(t *testing.T) {
	content := "addr: :8080\nupstream: http://localhost:9090\ncors:\n  enabled: true\n  allowed_origins: []\n"
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for cors enabled with no allowed_origins")
	}
}

func TestLoad_CORSValid(t *testing.T) {
	content := "addr: :8080\nupstream: http://localhost:9090\ncors:\n  enabled: true\n  allowed_origins:\n    - https://portal.gov\n"
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.CORS.Enabled {
		t.Error("expected CORS to be enabled")
	}
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "https://portal.gov" {
		t.Errorf("unexpected allowed origins: %v", cfg.CORS.AllowedOrigins)
	}
}
