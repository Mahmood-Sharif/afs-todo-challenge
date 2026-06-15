package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port     string
	Database DatabaseConfig
	Auth     AuthConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret      string
	JWTExpiryHours int
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
		Auth: AuthConfig{
			JWTSecret: os.Getenv("JWT_SECRET"),
		},
	}

	expiryHours, err := parsePositiveInt("JWT_EXPIRY_HOURS", os.Getenv("JWT_EXPIRY_HOURS"))
	if err != nil {
		return Config{}, err
	}
	cfg.Auth.JWTExpiryHours = expiryHours

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
		{key: "JWT_SECRET", value: cfg.Auth.JWTSecret},
		{key: "JWT_EXPIRY_HOURS", value: fmt.Sprintf("%d", cfg.Auth.JWTExpiryHours)},
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

func parsePositiveInt(key string, value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("%s is required", key)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", key)
	}

	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return parsed, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
