package vision

import (
	"fmt"
	"time"

	"github.com/echo404s/PicMind/backend/pkg/config"
)

// Config is vision-service process configuration.
type Config struct {
	Server  Server        `mapstructure:"server"`
	Consul  config.Consul `mapstructure:"consul"`
	Service Service       `mapstructure:"service"`
	Redis   Redis         `mapstructure:"redis"`
	Vision  VisionLLM     `mapstructure:"vision"`
}

// Server holds gRPC listen settings.
type Server struct {
	Port int `mapstructure:"port"`
}

// Service holds Consul registration identity.
type Service struct {
	Name    string   `mapstructure:"name"`
	ID      string   `mapstructure:"id"`
	Address string   `mapstructure:"address"`
	Tags    []string `mapstructure:"tags"`
}

// Redis holds cache / lock connection settings.
type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// VisionLLM holds Vision-LLM providers and call policy.
//
// Providers is an ordered failover chain: index 0 is primary, then 1..N.
type VisionLLM struct {
	Providers     []ModelProvider `mapstructure:"providers"`
	PromptVersion string          `mapstructure:"prompt_version"`
	Timeout       string          `mapstructure:"timeout"` // e.g. "60s"
	MaxConcurrent int             `mapstructure:"max_concurrent"`
	RateLimit     RateLimit       `mapstructure:"rate_limit"`
	Retry         Retry           `mapstructure:"retry"`
	Cache         Cache           `mapstructure:"cache"`
}

// ModelProvider describes one Vision-LLM endpoint.
type ModelProvider struct {
	Name    string `mapstructure:"name"`     // logical name, e.g. qwen-vl
	BaseURL string `mapstructure:"base_url"` // API base URL
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"` // model id sent to the provider
}

// RateLimit is a token-bucket limiter for outbound LLM calls.
type RateLimit struct {
	RPS   float64 `mapstructure:"rps"`
	Burst int     `mapstructure:"burst"`
}

// Retry configures exponential backoff for LLM calls.
type Retry struct {
	MaxAttempts    int    `mapstructure:"max_attempts"`
	InitialBackoff string `mapstructure:"initial_backoff"` // e.g. "200ms"
	MaxBackoff     string `mapstructure:"max_backoff"`     // e.g. "5s"
}

// Cache configures Redis result cache (key: md5 + model + prompt_version).
type Cache struct {
	Enabled bool   `mapstructure:"enabled"`
	TTL     string `mapstructure:"ttl"` // e.g. "24h"
}

// LoadConfig loads and unmarshals vision-service config from path.
func LoadConfig(path string) (*Config, error) {
	v, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal vision config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be > 0")
	}
	if c.Consul.Address == "" {
		return fmt.Errorf("consul.address is required")
	}
	if c.Service.Name == "" {
		return fmt.Errorf("service.name is required")
	}
	if c.Service.ID == "" {
		return fmt.Errorf("service.id is required")
	}
	if c.Service.Address == "" {
		return fmt.Errorf("service.address is required")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr is required")
	}
	if len(c.Vision.Providers) == 0 {
		return fmt.Errorf("vision.providers must contain at least one provider")
	}
	seen := make(map[string]struct{}, len(c.Vision.Providers))
	for i, p := range c.Vision.Providers {
		prefix := fmt.Sprintf("vision.providers[%d]", i)
		if err := p.validate(prefix); err != nil {
			return err
		}
		if _, ok := seen[p.Name]; ok {
			return fmt.Errorf("%s: duplicate provider name %q", prefix, p.Name)
		}
		seen[p.Name] = struct{}{}
	}
	if c.Vision.PromptVersion == "" {
		return fmt.Errorf("vision.prompt_version is required")
	}
	if _, err := parseDuration("vision.timeout", c.Vision.Timeout); err != nil {
		return err
	}
	if c.Vision.MaxConcurrent <= 0 {
		return fmt.Errorf("vision.max_concurrent must be > 0")
	}
	if c.Vision.RateLimit.RPS <= 0 {
		return fmt.Errorf("vision.rate_limit.rps must be > 0")
	}
	if c.Vision.RateLimit.Burst <= 0 {
		return fmt.Errorf("vision.rate_limit.burst must be > 0")
	}
	if c.Vision.Retry.MaxAttempts <= 0 {
		return fmt.Errorf("vision.retry.max_attempts must be > 0")
	}
	if _, err := parseDuration("vision.retry.initial_backoff", c.Vision.Retry.InitialBackoff); err != nil {
		return err
	}
	if _, err := parseDuration("vision.retry.max_backoff", c.Vision.Retry.MaxBackoff); err != nil {
		return err
	}
	if c.Vision.Cache.Enabled {
		if _, err := parseDuration("vision.cache.ttl", c.Vision.Cache.TTL); err != nil {
			return err
		}
	}
	return nil
}

func (m ModelProvider) validate(prefix string) error {
	if m.Name == "" {
		return fmt.Errorf("%s.name is required", prefix)
	}
	if m.BaseURL == "" {
		return fmt.Errorf("%s.base_url is required", prefix)
	}
	if m.Model == "" {
		return fmt.Errorf("%s.model is required", prefix)
	}
	return nil
}

func parseDuration(field, value string) (time.Duration, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required", field)
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid duration %q: %w", field, value, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s must be > 0", field)
	}
	return d, nil
}

// Primary returns the first provider in the failover chain.
func (v VisionLLM) Primary() ModelProvider {
	return v.Providers[0]
}

// TimeoutDuration returns the LLM request timeout.
func (v VisionLLM) TimeoutDuration() time.Duration {
	d, _ := time.ParseDuration(v.Timeout)
	return d
}

// InitialBackoffDuration returns the retry initial backoff.
func (r Retry) InitialBackoffDuration() time.Duration {
	d, _ := time.ParseDuration(r.InitialBackoff)
	return d
}

// MaxBackoffDuration returns the retry max backoff.
func (r Retry) MaxBackoffDuration() time.Duration {
	d, _ := time.ParseDuration(r.MaxBackoff)
	return d
}

// TTLDuration returns the cache TTL.
func (c Cache) TTLDuration() time.Duration {
	d, _ := time.ParseDuration(c.TTL)
	return d
}
