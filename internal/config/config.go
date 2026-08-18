package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr string `env:"SERVER_ADDRESS"`
	BaseUrl    string `env:"BASE_URL"`
}

func Load() (Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddr, "a", ":8080", "Server address")
	flag.StringVar(&cfg.BaseUrl, "b", "http://localhost:8080", "Base URL for short links redirection")

	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("parse env config: %w", err)
	}

	return cfg, nil
}
