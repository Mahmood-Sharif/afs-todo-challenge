package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port     string
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() (Config, error) {
	cfg := Config{
		Port: getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	missing := make([]string, 0)

	required := []struct {
		key   string
		value string
	}{
		{key: "DB_HOST", value: cfg.Database.Host},
		{key: "DB_PORT", value: cfg.Database.Port},
		{key: "DB_USER", value: cfg.Database.User},
		{key: "DB_PASSWORD", value: cfg.Database.Password},
		{key: "DB_NAME", value: cfg.Database.Name},
		{key: "DB_SSLMODE", value: cfg.Database.SSLMode},
	}

	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			missing = append(missing, item.key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
