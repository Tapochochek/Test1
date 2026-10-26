package config

import (
	"os"
	"strings"
)

const (
	DefaultPort     = "8080"
	DefaultDatabase = "postgres://orenkod:orenkod_local@localhost:5432/orenkod?sslmode=disable"
)

type Config struct {
	Port             string
	DatabaseURL      string
	MaxBotToken      string
	MaxWebhookSecret string
	MaxDefaultChatID string
	MaxAPIBaseURL    string
}

func Load() Config {
	return Config{
		Port:             env("PORT", DefaultPort),
		DatabaseURL:      env("DATABASE_URL", DefaultDatabase),
		MaxBotToken:      env("MAX_BOT_TOKEN", ""),
		MaxWebhookSecret: env("MAX_WEBHOOK_SECRET", ""),
		MaxDefaultChatID: env("MAX_DEFAULT_CHAT_ID", ""),
		MaxAPIBaseURL:    "https://platform-api2.max.ru",
	}
}

func (c Config) MaxConfigured() bool {
	return c.MaxBotToken != ""
}

func (c Config) DefaultChatConfigured() bool {
	return c.MaxDefaultChatID != ""
}

func (c Config) WebhookSecretConfigured() bool {
	return c.MaxWebhookSecret != ""
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
