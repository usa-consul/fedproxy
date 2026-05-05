package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level fedproxy configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Proxy    ProxyConfig    `yaml:"proxy"`
	SAML     SAMLConfig     `yaml:"saml"`
	PIV      PIVConfig      `yaml:"piv"`
}

// ServerConfig defines listener settings.
type ServerConfig struct {
	Addr            string        `yaml:"addr"`
	TLSCertFile     string        `yaml:"tls_cert_file"`
	TLSKeyFile      string        `yaml:"tls_key_file"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
}

// ProxyConfig defines upstream target settings.
type ProxyConfig struct {
	UpstreamURL     string        `yaml:"upstream_url"`
	FlushInterval   time.Duration `yaml:"flush_interval"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
}

// SAMLConfig holds SAML identity provider settings.
type SAMLConfig struct {
	Enabled         bool   `yaml:"enabled"`
	IDPMetadataURL  string `yaml:"idp_metadata_url"`
	EntityID        string `yaml:"entity_id"`
	ACSURL          string `yaml:"acs_url"`
	CertFile        string `yaml:"cert_file"`
	KeyFile         string `yaml:"key_file"`
}

// PIVConfig holds PIV/CAC smart-card authentication settings.
type PIVConfig struct {
	Enabled         bool     `yaml:"enabled"`
	ClientCAFiles   []string `yaml:"client_ca_files"`
	RequireClientCert bool   `yaml:"require_client_cert"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	cfg := &Config{}
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation: %w", err)
	}

	return cfg, nil
}

// validate performs basic sanity checks on the loaded configuration.
func (c *Config) validate() error {
	if c.Server.Addr == "" {
		return fmt.Errorf("server.addr must not be empty")
	}
	if c.Proxy.UpstreamURL == "" {
		return fmt.Errorf("proxy.upstream_url must not be empty")
	}
	if c.SAML.Enabled && c.SAML.IDPMetadataURL == "" {
		return fmt.Errorf("saml.idp_metadata_url required when saml is enabled")
	}
	if c.PIV.Enabled && len(c.PIV.ClientCAFiles) == 0 {
		return fmt.Errorf("piv.client_ca_files required when piv is enabled")
	}
	return nil
}
