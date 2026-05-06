package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// AuthMode defines the authentication strategy.
type AuthMode string

const (
	AuthNone AuthMode = "none"
	AuthSAML AuthMode = "saml"
	AuthPIV  AuthMode = "piv"
)

// CircuitBreakerConfig holds circuit breaker settings from config file.
type CircuitBreakerConfig struct {
	Enabled      bool          `yaml:"enabled"`
	MaxFailures  int           `yaml:"max_failures"`
	OpenDuration time.Duration `yaml:"open_duration"`
}

// Config represents the full application configuration.
type Config struct {
	Addr     string   `yaml:"addr"`
	Upstream string   `yaml:"upstream"`
	Auth     AuthMode `yaml:"auth"`

	SAMLMetadataURL string `yaml:"saml_metadata_url"`
	ExemptPaths     []string `yaml:"exempt_paths"`

	RateLimit struct {
		Enabled  bool `yaml:"enabled"`
		Requests int  `yaml:"requests"`
		Window   time.Duration `yaml:"window"`
	} `yaml:"rate_limit"`

	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`

	Timeout time.Duration `yaml:"timeout"`
	LogJSON bool          `yaml:"log_json"`
}

// Load reads and validates configuration from the given file path.
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
	if cfg.Auth == AuthSAML && cfg.SAMLMetadataURL == "" {
		return errors.New("config: saml_metadata_url required when auth is saml")
	}
	if cfg.CircuitBreaker.Enabled {
		if cfg.CircuitBreaker.MaxFailures <= 0 {
			return errors.New("config: circuit_breaker.max_failures must be > 0")
		}
		if cfg.CircuitBreaker.OpenDuration <= 0 {
			return errors.New("config: circuit_breaker.open_duration must be > 0")
		}
	}
	return nil
}
