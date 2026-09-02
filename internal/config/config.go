package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr      string `env:"SERVER_ADDRESS"`
	BaseUrl         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func Load() (Config, error) {
	var cfg Config

	flag.StringVar(&cfg.ServerAddr, "a", ":8080", "Server address")
	flag.StringVar(&cfg.BaseUrl, "b", "http://localhost:8080", "Base URL for short links redirection")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "Path for file storage")
	flag.StringVar(&cfg.DatabaseDSN, "d", "host=localhost user=postgres password=password dbname=postgres sslmode=disable", "Database dsn")

	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("parse env config: %w", err)
	}

	return cfg, nil
}
