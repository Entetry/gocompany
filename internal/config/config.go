// Package config contains app config files
package config

import (
	"github.com/caarlos0/env/v6"
)

// Config Main application config
type Config struct {
	Port             int    `env:"APP_PORT" envDefault:"22800"`
	ConnectionString string `env:"CONNECTION_STRING"`
	RedisPort        int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisHost        string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPass        string `env:"REDIS_PASS" envDefault:""`
}

// New Creates Config object
func New() (*Config, error) {
	cfg := new(Config)
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
