package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("MAX_BOT_TOKEN", "")
	t.Setenv("MAX_WEBHOOK_SECRET", "")
	t.Setenv("MAX_DEFAULT_CHAT_ID", "")

	cfg := Load()
	if cfg.Port != DefaultPort {
		t.Fatalf("port: got %q want %q", cfg.Port, DefaultPort)
	}
	if cfg.DatabaseURL != DefaultDatabase {
		t.Fatalf("database url mismatch")
	}
	if cfg.MaxConfigured() || cfg.WebhookSecretConfigured() || cfg.DefaultChatConfigured() {
		t.Fatal("MAX flags must be false when env is empty")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://demo")
	t.Setenv("MAX_BOT_TOKEN", " token ")
	t.Setenv("MAX_WEBHOOK_SECRET", "secret")
	t.Setenv("MAX_DEFAULT_CHAT_ID", "123")

	cfg := Load()
	if cfg.Port != "9090" {
		t.Fatalf("port: got %q", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://demo" {
		t.Fatalf("database: got %q", cfg.DatabaseURL)
	}
	if cfg.MaxBotToken != "token" {
		t.Fatalf("token should be trimmed, got %q", cfg.MaxBotToken)
	}
	if !cfg.MaxConfigured() || !cfg.WebhookSecretConfigured() || !cfg.DefaultChatConfigured() {
		t.Fatal("MAX flags must be true when env is set")
	}
}
