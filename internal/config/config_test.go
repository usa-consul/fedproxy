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
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
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
	path := writeTempConfig(t, "addr: :8080\nupstream: http://localhost:9090\nauth: saml\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing saml_metadata_url")
	}
}

func TestLoad_CircuitBreakerValid(t *testing.T) {
	content := `
addr: :8080
upstream: http://localhost:9090
circuit_breaker:
  enabled: true
  max_failures: 5
  open_duration: 30s
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.CircuitBreaker.Enabled {
		t.Error("expected circuit breaker enabled")
	}
	if cfg.CircuitBreaker.MaxFailures != 5 {
		t.Errorf("expected max_failures 5, got %d", cfg.CircuitBreaker.MaxFailures)
	}
	if cfg.CircuitBreaker.OpenDuration != 30*time.Second {
		t.Errorf("expected open_duration 30s, got %v", cfg.CircuitBreaker.OpenDuration)
	}
}

func TestLoad_CircuitBreakerInvalidMaxFailures(t *testing.T) {
	content := `
addr: :8080
upstream: http://localhost:9090
circuit_breaker:
  enabled: true
  max_failures: 0
  open_duration: 30s
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for max_failures <= 0")
	}
}

func TestLoad_CircuitBreakerInvalidOpenDuration(t *testing.T) {
	content := `
addr: :8080
upstream: http://localhost:9090
circuit_breaker:
  enabled: true
  max_failures: 3
  open_duration: 0s
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for open_duration <= 0")
	}
}
