package gpt

import (
	"context"
	"sync"

	"github.com/sashabaranov/go-openai"
)

type Service struct {
	client        *openai.Client
	maxTokens     int
	enableContext bool
	userMap       map[string]*userMessage
	lock          sync.Mutex
}

func NewService(config Config, maxTokens int, enableContext bool) *Service {
	client := openai.NewClient(config.Token)
	return &Service{
		client:        client,
		maxTokens:     maxTokens,
		enableContext: enableContext,
		userMap:       map[string]*userMessage{},
	}
}

func (s *Service) getUser(userIdentifier string) *userMessage {
	s.lock.Lock()
	defer s.lock.Unlock()
	user, ok := s.userMap[userIdentifier]
	if !ok {
		user = &userMessage{}
		s.userMap[userIdentifier] = user
	}
	return user
}

func (s *Service) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.userMap = map[string]*userMessage{}
}

func (s *Service) SendChatCompletion(userIdentifier, content string) (string, error) {
	user := s.getUser(userIdentifier)
	if s.enableContext {
		s.lock.Lock()
		user.cal(calTokens(content), s.maxTokens)
		s.lock.Unlock()
	}
	messages := append(user.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
			User:     userIdentifier,
		},
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	respContent := resp.Choices[0].Message.Content
	if s.enableContext {
		s.lock.Lock()
		defer s.lock.Unlock()
		user.messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: respContent,
		})
		user.tokens = append(user.tokens, resp.Usage.PromptTokens-user.totalToken, resp.Usage.CompletionTokens)
		user.totalToken = resp.Usage.TotalTokens
	}
	return respContent, nil
}

func calTokens(content string) int {
	return len(content) / 4
}
