package bot

import (
	"context"
	"fmt"
	"log"

	"math/rand"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-ai-assistant/ai"
	"telegram-ai-assistant/storage"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	ai      *ai.Service
	storage *storage.MemoryStorage
	mu      sync.RWMutex
	options map[string]string
}

func NewBot(token string, aiService *ai.Service, storage *storage.MemoryStorage) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{
		api:     api,
		ai:      aiService,
		storage: storage,
		options: make(map[string]string),
	}, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (b *Bot) saveOption(text string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	id := fmt.Sprintf("opt_%d", rand.Intn(1000000))
	b.options[id] = text
	return id
}

func (b *Bot) getOption(id string) (string, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	text, exists := b.options[id]
	return text, exists
}

func (b *Bot) Start() {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			go b.handleCallback(update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}
		
		go b.handleMessage(update.Message)
	}
}

func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	data := query.Data

	// Acknowledge the callback
	callback := tgbotapi.NewCallback(query.ID, "")
	b.api.Request(callback)

	text := data
	if strings.HasPrefix(data, "opt_") {
		if optText, ok := b.getOption(data); ok {
			text = optText
		} else {
			text = "Извините, этот вариант устарел. Пожалуйста, задайте вопрос заново."
			b.api.Send(tgbotapi.NewMessage(chatID, text))
			return
		}
	}

	// Static interceptions for standard services to save AI quota
	if text == "Расскажи подробнее про продажу лакокрасочной продукции" {
		replyText := "Мы являемся интернет-магазином строительных красок, лаков и малярных инструментов. В нашем ассортименте продукция ведущих мировых брендов: Dulux, Marshall, Hammerite, Pinotex, Master Color, Oikos, Maitre Deco."
		b.api.Send(tgbotapi.NewMessage(chatID, replyText))
		return
	}
	if text == "Расскажи подробнее про компьютерную колеровку" {
		replyText := "Мы предлагаем компьютерную колеровку красок. Используя передовое колеровочное оборудование, мы можем точно подобрать цвет из более чем 125 000 оттенков!"
		b.api.Send(tgbotapi.NewMessage(chatID, replyText))
		return
	}
	if text == "Расскажи подробнее про подбор материалов для интерьеров" {
		replyText := "Наши специалисты помогут вам с подбором идеальных материалов для вашего интерьера. Мы работаем как с розничными покупателями (B2C), так и с дизайнерами интерьеров и строительными бригадами."
		b.api.Send(tgbotapi.NewMessage(chatID, replyText))
		return
	}

	history := b.storage.GetHistory(chatID)

	ctx := context.Background()
	response, err := b.ai.GenerateResponse(ctx, history, text)
	if err != nil {
		log.Printf("[Error] AI failed for chat %d: %v", chatID, err)
		reply := tgbotapi.NewMessage(chatID, "Извините, сервис временно недоступен. Попробуйте позже.")
		b.api.Send(reply)
		return
	}

	b.storage.AddMessage(chatID, "Пользователь: "+text)
	b.storage.AddMessage(chatID, "Ассистент: "+response)

	b.sendDynamicResponse(chatID, response)
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text

	if text == "" {
		return
	}

	if msg.IsCommand() {
		if msg.Command() == "start" {
			reply := tgbotapi.NewMessage(chatID, "Здравствуйте! Я AI-ассистент компании «Центр красок». Выберите вопрос ниже или напишите свой ⬇️")
			reply.ReplyMarkup = getKeyboard()
			b.api.Send(reply)
		}
		return
	}

	if text == "🎨 Наши услуги" {
		replyText := "«Центр красок» предоставляет следующие услуги:\n\nВыберите интересующую вас услугу:"
		reply := tgbotapi.NewMessage(chatID, replyText)
		
		id1 := b.saveOption("Расскажи подробнее про продажу лакокрасочной продукции")
		id2 := b.saveOption("Расскажи подробнее про компьютерную колеровку")
		id3 := b.saveOption("Расскажи подробнее про подбор материалов для интерьеров")

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🛒 Продажа продукции", id1),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🎨 Компьютерная колеровка", id2),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Подбор материалов", id3),
			),
		)
		reply.ReplyMarkup = inlineKeyboard
		if _, err := b.api.Send(reply); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
		return
	}

	if text == "📍 Адреса магазинов" {
		replyText := "Наши магазины расположены по следующим адресам:\n\n• Алматы: ул. Желтоксан (точный адрес уточняйте по тел. +7 (727) 227-50-00)\n• Астана: Проспект Мангилик Ел, 29/2 (1 этаж)\n• Караганда: 137-й учетный квартал, 39"
		reply := tgbotapi.NewMessage(chatID, replyText)
		reply.ReplyMarkup = getKeyboard()
		b.api.Send(reply)
		return
	}

	if text == "🏷 Какие бренды есть?" {
		replyText := "Мы являемся официальным дистрибьютором следующих европейских и мировых брендов:\n\n• Dulux\n• Marshall\n• Hammerite\n• Pinotex\n• Master Color\n• Oikos\n• Maitre Deco"
		reply := tgbotapi.NewMessage(chatID, replyText)
		reply.ReplyMarkup = getKeyboard()
		b.api.Send(reply)
		return
	}

	if text == "📞 Как связаться?" {
		replyText := "Наши контакты:\n\n• Телефон (Алматы): +7 (727) 227-50-00\n• Email: dulux_armada@abis.kz\n• Режим работы (Алматы): Пн-Вс: 10:00 - 20:00"
		reply := tgbotapi.NewMessage(chatID, replyText)
		reply.ReplyMarkup = getKeyboard()
		b.api.Send(reply)
		return
	}

	history := b.storage.GetHistory(chatID)

	ctx := context.Background()
	response, err := b.ai.GenerateResponse(ctx, history, text)
	if err != nil {
		log.Printf("[Error] AI failed for chat %d: %v", chatID, err)
		reply := tgbotapi.NewMessage(chatID, "Извините, сервис временно недоступен. Попробуйте позже.")
		b.api.Send(reply)
		return
	}

	b.storage.AddMessage(chatID, "Пользователь: "+text)
	b.storage.AddMessage(chatID, "Ассистент: "+response)

	b.sendDynamicResponse(chatID, response)
}

func (b *Bot) sendDynamicResponse(chatID int64, text string) {
	lines := strings.Split(text, "\n")
	var cleanLines []string
	var options []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[ОПЦИЯ] ") {
			opt := strings.TrimPrefix(trimmed, "[ОПЦИЯ] ")
			// Clean up any bold tags for the button text
			opt = strings.ReplaceAll(opt, "**", "")
			options = append(options, opt)
		} else {
			cleanLines = append(cleanLines, line)
		}
	}

	finalText := strings.TrimSpace(strings.Join(cleanLines, "\n"))
	if finalText == "" && len(options) > 0 {
		finalText = "Выберите один из вариантов:"
	} else if finalText == "" {
		finalText = "..."
	}

	// Replace markdown bullet points with a standard bullet character
	finalText = strings.ReplaceAll(finalText, "* ", "• ")
	// Remove all remaining asterisks (usually used for **bold** markdown)
	finalText = strings.ReplaceAll(finalText, "*", "")

	reply := tgbotapi.NewMessage(chatID, finalText)

	if len(options) > 0 {
		var rows [][]tgbotapi.InlineKeyboardButton
		for _, opt := range options {
			id := b.saveOption(opt)
			
			// Shorten the display text if it's too long
			btnText := opt
			if len([]rune(btnText)) > 40 {
				btnText = string([]rune(btnText)[:37]) + "..."
			}
			
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(btnText, id),
			))
		}
		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	} else {
		reply.ReplyMarkup = getKeyboard()
	}

	if _, err := b.api.Send(reply); err != nil {
		log.Printf("Failed to send dynamic response: %v", err)
	}
}

func getKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎨 Наши услуги"),
			tgbotapi.NewKeyboardButton("📍 Адреса магазинов"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏷 Какие бренды есть?"),
			tgbotapi.NewKeyboardButton("📞 Как связаться?"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}
