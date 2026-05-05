package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the full fedproxy configuration.
type Config struct {
	Addr     string      `yaml:"addr"`
	Upstream string      `yaml:"upstream"`
	Auth     AuthConfig  `yaml:"auth"`
	Rate     RateConfig  `yaml:"rate"`
	Log      LogConfig   `yaml:"log"`
}

// AuthConfig controls authentication mode and related settings.
type AuthConfig struct {
	Mode         string   `yaml:"mode"` // none | saml | piv
	MetadataURL  string   `yaml:"metadata_url"`
	ExemptPaths  []string `yaml:"exempt_paths"`
}

// RateConfig controls per-IP rate limiting.
type RateConfig struct {
	Enabled       bool `yaml:"enabled"`
	RequestsPerMin int  `yaml:"requests_per_min"`
}

// LogConfig controls request logging behaviour.
type LogConfig struct {
	Format string `yaml:"format"` // text | json
}

// Load reads and validates a YAML config file at the given path.
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
	if cfg.Auth.Mode == "saml" && cfg.Auth.MetadataURL == "" {
		return errors.New("config: auth.metadata_url required when mode is saml")
	}
	if cfg.Rate.Enabled && cfg.Rate.RequestsPerMin <= 0 {
		return errors.New("config: rate.requests_per_min must be > 0 when rate limiting is enabled")
	}
	return nil
}
