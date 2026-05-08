package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for fedproxy.
type Config struct {
	Addr     string        `yaml:"addr"`
	Upstream string        `yaml:"upstream"`
	Timeout  time.Duration `yaml:"timeout"`

	Auth AuthConfig `yaml:"auth"`

	RateLimit RateLimitConfig `yaml:"rate_limit"`

	CORS CORSConfig `yaml:"cors"`
}

// AuthConfig controls authentication mode.
type AuthConfig struct {
	Mode            string   `yaml:"mode"` // none | saml | piv
	SAMLMetadata    string   `yaml:"saml_metadata"`
	ExemptPaths     []string `yaml:"exempt_paths"`
}

// RateLimitConfig controls per-client rate limiting.
type RateLimitConfig struct {
	Enabled  bool `yaml:"enabled"`
	Requests int  `yaml:"requests"`
	Window   time.Duration `yaml:"window"`
}

// CORSConfig mirrors middleware.CORSConfig for YAML deserialization.
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           string   `yaml:"max_age"`
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
	if cfg.Auth.Mode == "saml" && cfg.Auth.SAMLMetadata == "" {
		return errors.New("config: saml_metadata is required when auth mode is saml")
	}
	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.Requests <= 0 {
			return errors.New("config: rate_limit.requests must be > 0 when rate limiting is enabled")
		}
		if cfg.RateLimit.Window <= 0 {
			return errors.New("config: rate_limit.window must be > 0 when rate limiting is enabled")
		}
	}
	if cfg.CORS.Enabled && len(cfg.CORS.AllowedOrigins) == 0 {
		return errors.New("config: cors.allowed_origins must not be empty when CORS is enabled")
	}
	return nil
}
