package ernie

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/sashabaranov/go-openai"
	"github.com/vanus-labs/vanus-connect-runtime/internal/apiserver/handlers/chatai/ernie/oauth"
	"golang.org/x/oauth2"
)

const url = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"

type Service struct {
	client        *resty.Client
	tokenSource   oauth2.TokenSource
	config        Config
	maxTokens     int
	enableContext bool
	userMap       map[string]*userMessage
	lock          sync.Mutex
}

func NewService(config Config, maxTokens int, enableContext bool) *Service {
	return &Service{
		config:        config,
		maxTokens:     1500,
		enableContext: enableContext,
		userMap:       map[string]*userMessage{},
		client:        resty.New(),
		tokenSource:   oauth.NewTokenSource(config.AccessKey, config.SecretKey),
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
	token, err := s.tokenSource.Token()
	if err != nil {
		return "", err
	}
	user := s.getUser(userIdentifier)
	if s.enableContext {
		s.lock.Lock()
		user.cal(calTokens(content), s.maxTokens)
		s.lock.Unlock()
	}
	messages := append(user.messages, ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})
	req := ChatCompletionRequest{
		Message: messages,
		User:    userIdentifier,
	}
	res, err := s.client.R().SetQueryParam("access_token", token.AccessToken).
		SetHeader("Content-Type", "application/json").SetBody(req).Post(url)
	if err != nil {
		return "", err
	}
	var resp ChatCompletionResponse
	err = json.Unmarshal(res.Body(), &resp)
	if err != nil {
		return "", err
	}
	if resp.ErrorCode != 0 {
		return "", fmt.Errorf("response error code:%d, msg:%s", resp.ErrorCode, resp.ErrorMsg)
	}
	respContent := resp.Result
	if s.enableContext {
		s.lock.Lock()
		defer s.lock.Unlock()
		user.messages = append(messages, ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: respContent,
		})
		user.tokens = append(user.tokens, resp.Usage.PromptTokens-user.totalToken, resp.Usage.CompletionTokens)
		user.totalToken = resp.Usage.TotalTokens
	}
	return respContent, nil
}

func calTokens(content string) int {
	return len(content)
}
