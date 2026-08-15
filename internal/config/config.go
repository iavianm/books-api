package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPPort string
	CacheTTL time.Duration
	DB       DBConfig
}

type DBConfig struct {
	Host, Port, User, Password, Name, SSLMode string
}

func LoadConfig() *Config {
	return &Config{
		HTTPPort: mustEnv("HTTP_PORT"),
		CacheTTL: mustDuration("CACHE_TTL"),
		DB: DBConfig{
			Host:     mustEnv("DB_HOST"),
			Port:     mustEnv("DB_PORT"),
			User:     mustEnv("DB_USER"),
			Password: mustEnv("DB_PASSWORD"),
			Name:     mustEnv("DB_NAME"),
			SSLMode:  mustEnv("DB_SSLMODE"),
		},
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("%s environment variable not set", key))
	}
	return v
}

func mustDuration(key string) time.Duration {
	d, err := time.ParseDuration(mustEnv(key))
	if err != nil {
		panic(fmt.Sprintf("%s must be a duration like 30s: %v", key, err))
	}
	if d <= 0 {
		panic(fmt.Sprintf("%s must be positive", key))
	}
	return d
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}
