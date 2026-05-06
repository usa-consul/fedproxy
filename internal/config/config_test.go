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
		t.Errorf("addr = %q, want :8080", cfg.Addr)
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

func TestLoad_PIVMissingCACert(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\nauth:\n  mode: piv\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for piv mode without ca cert")
	}
}

func TestLoad_SecuritySection(t *testing.T) {
	content := `
addr: :8080
upstream: http://localhost:9090
security:
  hsts_max_age: 63072000
  content_security_policy: "default-src 'self'"
  extra_headers:
    X-Agency: DOD
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Security.HSTSMaxAge != 63072000 {
		t.Errorf("hsts_max_age = %d, want 63072000", cfg.Security.HSTSMaxAge)
	}
	if cfg.Security.ExtraHeaders["X-Agency"] != "DOD" {
		t.Errorf("extra header X-Agency = %q, want DOD", cfg.Security.ExtraHeaders["X-Agency"])
	}
}

func TestLoad_RateLimitMissingRequests(t *testing.T) {
	content := "addr: :8080\nupstream: http://localhost:9090\nrate_limit:\n  enabled: true\n  window_secs: 60\n"
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for rate_limit enabled without requests")
	}
}
