package config

import (
	"strings"
	"testing"
	"time"
)

func setAllEnv(t *testing.T) {
	t.Helper()

	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("CACHE_TTL", "30s")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres_user")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "books")
	t.Setenv("DB_SSLMODE", "disable")
}

func TestLoadConfig(t *testing.T) {
	setAllEnv(t)

	conf := LoadConfig()

	if conf.HTTPPort != "8080" {
		t.Errorf("HTTPPort = %q", conf.HTTPPort)
	}
	if conf.CacheTTL != 30*time.Second {
		t.Errorf("CacheTTL = %v, want 30s", conf.CacheTTL)
	}
	if conf.DB.Host != "localhost" || conf.DB.Port != "5432" {
		t.Errorf("DB = %+v", conf.DB)
	}
	if conf.DB.User != "postgres_user" || conf.DB.Password != "secret" {
		t.Errorf("credentials = %+v", conf.DB)
	}
	if conf.DB.Name != "books" || conf.DB.SSLMode != "disable" {
		t.Errorf("DB = %+v", conf.DB)
	}
}

func TestLoadConfigPanicsOnMissingVariable(t *testing.T) {
	required := []string{
		"HTTP_PORT", "CACHE_TTL", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
	}

	for _, key := range required {
		t.Run("missing "+key, func(t *testing.T) {
			setAllEnv(t)
			t.Setenv(key, "")

			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("expected panic when %s is unset", key)
				}
				msg, ok := r.(string)
				if !ok || !strings.Contains(msg, key) {
					t.Errorf("panic message = %v, want it to mention %s", r, key)
				}
			}()

			LoadConfig()
		})
	}
}

func TestLoadConfigPanicsOnInvalidCacheTTL(t *testing.T) {
	for _, value := range []string{"abc", "30", "-5s", "0s"} {
		t.Run(value, func(t *testing.T) {
			setAllEnv(t)
			t.Setenv("CACHE_TTL", value)

			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for CACHE_TTL=%q", value)
				}
			}()

			LoadConfig()
		})
	}
}

func TestDSN(t *testing.T) {
	db := DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres_user",
		Password: "secret",
		Name:     "books",
		SSLMode:  "disable",
	}

	dsn := db.DSN()

	for _, want := range []string{
		"host=localhost",
		"port=5432",
		"user=postgres_user",
		"password=secret",
		"dbname=books",
		"sslmode=disable",
	} {
		if !strings.Contains(dsn, want) {
			t.Errorf("DSN = %q, want it to contain %q", dsn, want)
		}
	}
}
