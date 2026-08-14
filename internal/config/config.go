package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPPort string
	DB       DBConfig
}

type DBConfig struct {
	Host, Port, User, Password, Name, SSLMode string
}

func LoadConfig() *Config {
	return &Config{
		HTTPPort: mustEnv("HTTP_PORT"),
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

func (c DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}
