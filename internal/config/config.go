package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the full fedproxy runtime configuration.
type Config struct {
	Addr     string `yaml:"addr"`
	Upstream string `yaml:"upstream"`

	Auth struct {
		Mode           string   `yaml:"mode"`            // none | saml | piv
		SAMLMetadata   string   `yaml:"saml_metadata"`   // path or URL
		ExemptPaths    []string `yaml:"exempt_paths"`
		PIVCACertFile  string   `yaml:"piv_ca_cert_file"`
	} `yaml:"auth"`

	RateLimit struct {
		Enabled    bool `yaml:"enabled"`
		Requests   int  `yaml:"requests"`
		WindowSecs int  `yaml:"window_secs"`
	} `yaml:"rate_limit"`

	Security struct {
		HSTSMaxAge            int               `yaml:"hsts_max_age"`
		ContentSecurityPolicy string            `yaml:"content_security_policy"`
		ExtraHeaders          map[string]string `yaml:"extra_headers"`
	} `yaml:"security"`
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
	switch cfg.Auth.Mode {
	case "", "none":
		cfg.Auth.Mode = "none"
	case "saml":
		if cfg.Auth.SAMLMetadata == "" {
			return errors.New("config: auth.saml_metadata is required when mode is saml")
		}
	case "piv":
		if cfg.Auth.PIVCACertFile == "" {
			return errors.New("config: auth.piv_ca_cert_file is required when mode is piv")
		}
	default:
		return errors.New("config: auth.mode must be one of: none, saml, piv")
	}
	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.Requests <= 0 {
			return errors.New("config: rate_limit.requests must be > 0 when rate limiting is enabled")
		}
		if cfg.RateLimit.WindowSecs <= 0 {
			return errors.New("config: rate_limit.window_secs must be > 0 when rate limiting is enabled")
		}
	}
	return nil
}
