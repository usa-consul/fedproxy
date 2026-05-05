package config

import (
	"os"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "fedproxy-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidMinimal(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.Addr)
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
		t.Fatal("expected error for saml mode without metadata_url")
	}
}

func TestLoad_RateEnabledZeroRequests(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\nrate:\n  enabled: true\n  requests_per_min: 0\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for rate enabled with zero requests_per_min")
	}
}

func TestLoad_ValidWithRate(t *testing.T) {
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\nrate:\n  enabled: true\n  requests_per_min: 60\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Rate.Enabled {
		t.Error("expected rate limiting to be enabled")
	}
	if cfg.Rate.RequestsPerMin != 60 {
		t.Errorf("expected 60, got %d", cfg.Rate.RequestsPerMin)
	}
}
