package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "fedproxy.yaml")
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("writeTempConfig: %v", err)
	}
	return p
}

func TestLoad_ValidMinimal(t *testing.T) {
	cfgYAML := `
server:
  addr: ":8443"
proxy:
  upstream_url: "http://localhost:9000"
`
	cfg, err := Load(writeTempConfig(t, cfgYAML))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Server.Addr != ":8443" {
		t.Errorf("server.addr = %q, want :8443", cfg.Server.Addr)
	}
	if cfg.Proxy.UpstreamURL != "http://localhost:9000" {
		t.Errorf("proxy.upstream_url = %q", cfg.Proxy.UpstreamURL)
	}
}

func TestLoad_MissingAddr(t *testing.T) {
	cfgYAML := `
proxy:
  upstream_url: "http://localhost:9000"
`
	_, err := Load(writeTempConfig(t, cfgYAML))
	if err == nil {
		t.Fatal("expected validation error for missing server.addr")
	}
}

func TestLoad_MissingUpstream(t *testing.T) {
	cfgYAML := `
server:
  addr: ":8443"
`
	_, err := Load(writeTempConfig(t, cfgYAML))
	if err == nil {
		t.Fatal("expected validation error for missing proxy.upstream_url")
	}
}

func TestLoad_SAMLEnabledMissingMetadata(t *testing.T) {
	cfgYAML := `
server:
  addr: ":8443"
proxy:
  upstream_url: "http://localhost:9000"
saml:
  enabled: true
`
	_, err := Load(writeTempConfig(t, cfgYAML))
	if err == nil {
		t.Fatal("expected validation error for saml enabled without idp_metadata_url")
	}
}

func TestLoad_PIVEnabledMissingCAs(t *testing.T) {
	cfgYAML := `
server:
  addr: ":8443"
proxy:
  upstream_url: "http://localhost:9000"
piv:
  enabled: true
`
	_, err := Load(writeTempConfig(t, cfgYAML))
	if err == nil {
		t.Fatal("expected validation error for piv enabled without client_ca_files")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/fedproxy.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
