package main

import (
	"context"
	"log"
	"os"

	"telegram-ai-assistant/ai"
	"telegram-ai-assistant/bot"
	"telegram-ai-assistant/config"
	"telegram-ai-assistant/storage"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	kbBytes, err := os.ReadFile("knowledge_base.txt")
	if err != nil {
		log.Fatalf("Failed to read knowledge base: %v", err)
	}
	knowledgeBase := string(kbBytes)

	ctx := context.Background()
	aiService, err := ai.NewService(ctx, cfg.GeminiAPIKey, knowledgeBase)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}
	defer aiService.Close()

	memStorage := storage.NewMemoryStorage(5)

	tgBot, err := bot.NewBot(cfg.TelegramToken, aiService, memStorage)
	if err != nil {
		log.Fatalf("Failed to start Telegram bot: %v", err)
	}

	log.Println("Starting Telegram bot...")
	tgBot.Start()
}
