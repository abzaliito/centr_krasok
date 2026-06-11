package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Service struct {
	client        *genai.Client
	model         *genai.GenerativeModel
	knowledgeBase string
}

func NewService(ctx context.Context, apiKey string, knowledgeBase string) (*Service, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash")
	
	systemInstruction := fmt.Sprintf(`Ты — профессиональный AI-ассистент компании 'Центр красок'. Твоя задача — отвечать на вопросы пользователей на основе предоставленной базы знаний. 
ПРАВИЛА И ОГРАНИЧЕНИЯ:
1. Отвечай исключительно на русском языке, вежливо и дружелюбно.
2. Используй ТОЛЬКО информацию из предоставленной базы знаний. 
3. Если ответа на вопрос пользователя нет в базе знаний, или если пользователь спрашивает о чем-то отвлеченном (не связанном с компанией), строго отвечай следующей фразой: 'К сожалению, у меня нет этой информации. Пожалуйста, свяжитесь с нашими специалистами по телефону +7 (727) 227-50-00, и они обязательно вам помогут.'
4. Категорически запрещено выдумывать цены, сроки, несуществующие акции, товары или вакансии.
5. ВАЖНО: Если ты предлагаешь пользователю выбрать один из вариантов для продолжения диалога (например, список услуг для уточнения), обязательно начинай каждую такую строку с префикса '[ОПЦИЯ] '. Обычные информационные списки (например, адреса или бренды) оформляй стандартными маркерами без этого префикса.

БАЗА ЗНАНИЙ:
%s`, knowledgeBase)

	model.SystemInstruction = genai.NewUserContent(genai.Text(systemInstruction))

	return &Service{
		client:        client,
		model:         model,
		knowledgeBase: knowledgeBase,
	}, nil
}

func (s *Service) GenerateResponse(ctx context.Context, history []string, query string) (string, error) {
	var promptBuilder strings.Builder
	if len(history) > 0 {
		promptBuilder.WriteString("Контекст последних сообщений:\n")
		for _, msg := range history {
			promptBuilder.WriteString(msg + "\n")
		}
		promptBuilder.WriteString("\nНовое сообщение пользователя: ")
	}
	promptBuilder.WriteString(query)
	
	resp, err := s.model.GenerateContent(ctx, genai.Text(promptBuilder.String()))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}
	
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from model")
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText += string(txt)
		}
	}

	return responseText, nil
}

func (s *Service) Close() {
	s.client.Close()
}
