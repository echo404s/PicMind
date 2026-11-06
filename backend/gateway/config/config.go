package config

import (
	"fmt"

	"github.com/echo404s/PicMind/backend/pkg/config"
)

// Config is api-gateway process configuration.
type Config struct {
	Server  Server        `mapstructure:"server"`
	Consul  config.Consul `mapstructure:"consul"`
	Service Service       `mapstructure:"service"`
}

// Server holds HTTP listen settings.
type Server struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// Service holds gateway identity metadata.
type Service struct {
	Name    string `mapstructure:"name"`
	ID      string `mapstructure:"id"`
	Address string `mapstructure:"address"`
}

// LoadConfig loads and unmarshals gateway config from path.
func LoadConfig(path string) (*Config, error) {
	v, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal gateway config: %w", err)
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
	if c.Server.Mode == "" {
		return fmt.Errorf("server.mode is required")
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
	return nil
}
