package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Consul is shared Consul connection settings.
type Consul struct {
	Address string `mapstructure:"address"`
	Token   string `mapstructure:"token"`
}

// Load reads a YAML config file at path and returns a viper instance.
func Load(path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	return v, nil
}
