package config

import (
	"strings"
	"testing"
	"time"
)

func TestDBStringHidesPassword(t *testing.T) {
	t.Parallel()

	db := DB{
		Host:          "localhost",
		Port:          5432,
		User:          "postgres",
		Password:      "secret",
		DBName:        "subaggregator",
		SSLMode:       "disable",
		MigrationPath: "migrations",
	}

	got := db.String()
	if strings.Contains(got, "secret") {
		t.Fatalf("DB.String() = %q, want hidden password", got)
	}

	if !strings.Contains(got, "Password: [HIDDEN]") {
		t.Fatalf("DB.String() = %q, want hidden password marker", got)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("API_HOST", "127.0.0.1")
	t.Setenv("API_PORT", "9090")
	t.Setenv("API_READ_TIMEOUT", "15s")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_IS_JSON", "false")
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "app")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "appdb")
	t.Setenv("DB_SSL_MODE", "require")
	t.Setenv("DB_MIGRATION_PATH", "custom-migrations")

	cfg, err := loadFromEnv()
	if err != nil {
		t.Fatalf("loadFromEnv() error = %v, want nil", err)
	}

	if cfg.API.Host != "127.0.0.1" {
		t.Errorf("loadFromEnv().API.Host = %q, want %q", cfg.API.Host, "127.0.0.1")
	}

	if cfg.API.Port != 9090 {
		t.Errorf("loadFromEnv().API.Port = %d, want %d", cfg.API.Port, 9090)
	}

	if cfg.API.ReadTimeout != 15*time.Second {
		t.Errorf("loadFromEnv().API.ReadTimeout = %s, want %s", cfg.API.ReadTimeout, 15*time.Second)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("loadFromEnv().Logging.Level = %q, want %q", cfg.Logging.Level, "debug")
	}

	if cfg.Logging.IsJSON {
		t.Error("loadFromEnv().Logging.IsJSON = true, want false")
	}

	if cfg.DB.Password != "secret" {
		t.Errorf("loadFromEnv().DB.Password = %q, want %q", cfg.DB.Password, "secret")
	}

	if cfg.DB.MigrationPath != "custom-migrations" {
		t.Errorf("loadFromEnv().DB.MigrationPath = %q, want %q", cfg.DB.MigrationPath, "custom-migrations")
	}
}
