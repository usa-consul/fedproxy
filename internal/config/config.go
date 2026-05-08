package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// RewriteRule mirrors middleware.RewriteRule for config parsing.
type RewriteRule struct {
	StripPrefix string `yaml:"strip_prefix"`
	AddPrefix   string `yaml:"add_prefix"`
}

// SAMLConfig holds SAML-specific settings.
type SAMLConfig struct {
	MetadataURL string `yaml:"metadata_url"`
}

// PIVConfig holds PIV/CAC certificate auth settings.
type PIVConfig struct {
	CABundle string `yaml:"ca_bundle"`
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	Burst             int     `yaml:"burst"`
}

// Config is the top-level application configuration.
type Config struct {
	Addr          string          `yaml:"addr"`
	Upstream      string          `yaml:"upstream"`
	AuthMode      string          `yaml:"auth_mode"`
	ExemptPaths   []string        `yaml:"exempt_paths"`
	SAML          SAMLConfig      `yaml:"saml"`
	PIV           PIVConfig       `yaml:"piv"`
	RateLimit     RateLimitConfig `yaml:"rate_limit"`
	RewriteRules  []RewriteRule   `yaml:"rewrite_rules"`
	ReadTimeout   int             `yaml:"read_timeout_seconds"`
	WriteTimeout  int             `yaml:"write_timeout_seconds"`
}

// Load reads a YAML config file from path and validates it.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Addr == "" {
		return errors.New("config: addr is required")
	}
	if cfg.Upstream == "" {
		return errors.New("config: upstream is required")
	}
	switch cfg.AuthMode {
	case "", "none", "piv":
		// valid
	case "saml":
		if cfg.SAML.MetadataURL == "" {
			return errors.New("config: saml.metadata_url is required when auth_mode is saml")
		}
	default:
		return errors.New("config: unknown auth_mode: " + cfg.AuthMode)
	}
	for _, rule := range cfg.RewriteRules {
		if rule.StripPrefix == "" && rule.AddPrefix == "" {
			return errors.New("config: rewrite rule must specify strip_prefix or add_prefix")
		}
	}
	return nil
}
