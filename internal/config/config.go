package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all runtime configuration for fedproxy.
type Config struct {
	Addr     string   `yaml:"addr"`
	Upstream string   `yaml:"upstream"`
	Exempt   []string `yaml:"exempt"`

	Auth struct {
		Mode string `yaml:"mode"` // none | saml | piv
	} `yaml:"auth"`

	SAML struct {
		MetadataURL string `yaml:"metadata_url"`
	} `yaml:"saml"`

	RateLimit struct {
		Requests int           `yaml:"requests"`
		Window   time.Duration `yaml:"window"`
	} `yaml:"rate_limit"`

	Timeout struct {
		Request time.Duration `yaml:"request"`
		Message string        `yaml:"message"`
	} `yaml:"timeout"`

	Security struct {
		HSTSMaxAge  int               `yaml:"hsts_max_age"`
		ExtraHeaders map[string]string `yaml:"extra_headers"`
	} `yaml:"security"`
}

// Load reads and validates a YAML config file at path.
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
	mode := cfg.Auth.Mode
	if mode != "" && mode != "none" && mode != "saml" && mode != "piv" {
		return errors.New("config: auth.mode must be none, saml, or piv")
	}
	if mode == "saml" && cfg.SAML.MetadataURL == "" {
		return errors.New("config: saml.metadata_url is required when auth.mode is saml")
	}
	if cfg.Timeout.Request < 0 {
		return errors.New("config: timeout.request must be non-negative")
	}
	if cfg.RateLimit.Requests < 0 {
		return errors.New("config: rate_limit.requests must be non-negative")
	}
	return nil
}
