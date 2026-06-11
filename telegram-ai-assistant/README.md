# Telegram AI Assistant - "Центр красок"

A production-ready Telegram MVP bot for "Центр красок" acting as a natural AI chat assistant.

## Features
- Natural conversation without inline menus or slash commands.
- Context-aware responses based on a thread-safe in-memory sliding window history limit (5 messages per chat).
- Google Generative AI (Gemini 1.5 Flash) integration with a strict prompt structure to prevent hallucinations.
- Graceful error handling and concurrency safe context management using `sync.RWMutex`.

## Requirements
- Go 1.20+
- Telegram Bot Token (from BotFather)
- Google Gemini API Key (from Google AI Studio)

## Setup

1. **Clone the repository and install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Configuration:**
   Copy `.env.example` to `.env` and fill in your tokens.
   ```bash
   cp .env.example .env
   ```

3. **Knowledge Base:**
   Ensure `knowledge_base.txt` contains up-to-date company data.

4. **Run the bot:**
   ```bash
   go run main.go
   ```

## Architecture
- `config/` - Loads environmental variables.
- `ai/` - Integrates with Google Generative AI SDK to process prompts.
- `storage/` - Handles the concurrent map logic with RWMutex for chat histories.
- `bot/` - Runs the Telegram bot logic, routing, and update loop.
