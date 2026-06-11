package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	GeminiAPIKey  string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load() // Ignore error as .env might not exist in production environment

	tgToken := os.Getenv("TELEGRAM_TOKEN")
	if tgToken == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	return &Config{
		TelegramToken: tgToken,
		GeminiAPIKey:  geminiKey,
	}, nil
}
