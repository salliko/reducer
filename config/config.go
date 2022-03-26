package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func (c *Config) Parse() error {
	if err := env.Parse(c); err != nil {
		return err
	}

	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "server address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "base url")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "file storage path")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "database dsn")

	flag.Parse()

	return nil
}
