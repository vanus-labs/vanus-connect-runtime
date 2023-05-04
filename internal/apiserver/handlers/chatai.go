// Copyright 2023 Linkall Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/google/uuid"

	ce "github.com/cloudevents/sdk-go/v2"
	sdkgo "github.com/vanus-labs/sdk/golang"
	"github.com/vanus-labs/vanus-connect-runtime/api/models"
	"github.com/vanus-labs/vanus-connect-runtime/api/restapi/operations/connector"
	"github.com/vanus-labs/vanus-connect-runtime/internal/apiserver/handlers/chatai/ernie"
	"github.com/vanus-labs/vanus-connect-runtime/internal/apiserver/handlers/chatai/gpt"
	"github.com/vanus-labs/vanus-connect-runtime/internal/apiserver/utils"
	"gopkg.in/yaml.v2"
	log "k8s.io/klog/v2"
)

// All registered processing functions should appear under Registxxx in order
func RegistChatAIHandler(a *Api) {
	a.ConnectorChataiHandler = connector.ChataiHandlerFunc(a.chataiHandler)
}

func (a *Api) chataiHandler(params connector.ChataiParams) middleware.Responder {
	c, err := a.ctrl().ControllersLister().Get(fmt.Sprintf("vanus/source-chatai-%s", params.ConnectorID))
	if err != nil {
		log.Warningf("failed to get connector: %s\n", params.ConnectorID)
		return utils.Response(500, err)
	}

	chataiConfig := &ChatAIConfig{}
	err = yaml.Unmarshal([]byte(c.Spec.Config), chataiConfig)
	if err != nil {
		return utils.Response(500, err)
	}

	chataiConfig.init()
	eventSource := params.HTTPRequest.Header.Get(headerSource)
	if eventSource == "" {
		eventSource = defaultChatAIEventSource
	}
	eventType := params.HTTPRequest.Header.Get(headerType)
	if eventType == "" {
		eventType = defaultChatAIEventType
	}

	// parse config
	u, err := url.Parse(chataiConfig.Target)
	if err != nil {
		log.Errorf("failed to parse target, connector_id: %s, err: %+v\n", params.ConnectorID, err)
		return utils.Response(500, err)
	}
	if ChatAIClient == nil {
		// use sdk publish event
		opts := &sdkgo.ClientOptions{
			// TODO(jiangkai): use target host
			Endpoint: "vanus-gateway.vanus:8080",
			Token:    "admin",
		}

		ChatAIClient, err = sdkgo.Connect(opts)
		if err != nil {
			log.Errorf("failed to connect to Vanus cluster, err: %+v\n", err)
			return utils.Response(500, err)
		}
	}
	subPaths := strings.Split(u.Path, "/")
	p := ChatAIClient.Publisher(sdkgo.WithEventbus(subPaths[2], subPaths[4]))

	var chatType Type
	chatMode := params.HTTPRequest.Header.Get(headerChatMode)
	if chatMode == "" {
		chatMode = params.HTTPRequest.Header.Get(headerChatModeOld)
	}
	if chatMode != "" {
		chatType = Type(chatMode)
		switch chatType {
		case ChatGPT, ChatErnieBot:
		default:
			return utils.Response(500, errors.New("chat_mode invalid"))
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	var userIdentifier string
	if chataiConfig.UserIdentifierHeader != "" {
		userIdentifier = params.HTTPRequest.Header.Get(chataiConfig.UserIdentifierHeader)
		if userIdentifier == "" {
			return utils.Response(500, errors.New("header userIdentifier is empty"))
		}
	}
	if ChatAIServer == nil {
		ChatAIServer = NewChatAIService(chataiConfig)
	}
	go func(params connector.ChataiParams, userIdentifier string) {
		defer wg.Done()
		event := ce.NewEvent()
		event.SetID(uuid.NewString())
		event.SetTime(time.Now())
		event.SetType(eventType)
		event.SetSource(eventSource)
		content, err := ChatAIServer.ChatCompletion(chatType, userIdentifier, params.Message)
		if err != nil {
			log.Warning("failed to get content from %s Chat: %+v\n", chatType, err)
		}
		log.Infof("get content from ChatAI success, content: %s\n", content)
		event.SetData(ce.ApplicationJSON, map[string]string{
			"result": content,
		})
		err = p.Publish(context.Background(), &event)
		if err != nil {
			log.Errorf("publish event failed, err: %s\n", err.Error())
			return
		}
	}(params, userIdentifier)
	if ChatAIServer.isSync(params.HTTPRequest) {
		wg.Wait()
	}
	return connector.NewChataiOK().WithPayload(&models.APIResponse{
		Code:    200,
		Message: "success",
	})
}

func (s *ChatAIService) isSync(req *http.Request) bool {
	processMode := req.Header.Get(headerProcessMode)
	if processMode == "" {
		processMode = s.config.DefaultProcessMode
	}
	return processMode == "sync"
}

var (
	ChatAIServer *ChatAIService
	ChatAIClient sdkgo.Client
)

type Type string

const (
	ChatGPT      Type = "chatgpt"
	ChatErnieBot Type = "wenxin"
)

type Auth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

func (a *Auth) IsEmpty() bool {
	return a == nil || a.Username == "" || a.Password == ""
}

type ChatAIConfig struct {
	Port                 int          `json:"port" yaml:"port"`
	Target               string       `json:"target" yaml:"target"`
	MaxTokens            int          `json:"max_tokens" yaml:"max_tokens"`
	EverydayLimit        int          `json:"everyday_limit" yaml:"everyday_limit"`
	EnableContext        bool         `json:"enable_context" yaml:"enable_context"`
	DefaultProcessMode   string       `json:"default_process_mode" yaml:"default_process_mode"`
	UserIdentifierHeader string       `json:"user_identifier_header" yaml:"user_identifier_header"`
	DefaultChatMode      Type         `json:"default_chat_mode" yaml:"default_chat_mode"`
	GPT                  gpt.Config   `json:"gpt" yaml:"gpt"`
	ErnieBot             ernie.Config `json:"ernie_bot" yaml:"ernie_bot"`
	Auth                 *Auth        `json:"auth" yaml:"auth"`
}

func (c *ChatAIConfig) init() {
	if c.DefaultChatMode == "" {
		c.DefaultChatMode = ChatGPT
	}
	if c.EverydayLimit <= 0 {
		c.EverydayLimit = 1000
	}
	if c.MaxTokens <= 0 {
		c.MaxTokens = 3500
	}
}

const (
	defaultChatAIEventType   = "vanus-chatAI-type"
	defaultChatAIEventSource = "vanus-chatAI-source"
	headerContentType        = "Content-Type"
	headerChatModeOld        = "Chat_Mode"
	headerChatMode           = "Chat-Mode"
	headerProcessMode        = "Process-Mode"
	applicationJSON          = "application/json"
)

type ChatAIService struct {
	chatGpt      ChatAIClientInterfacer
	ernieBot     ChatAIClientInterfacer
	config       *ChatAIConfig
	lock         sync.RWMutex
	day          int
	limitContent string
	userNum      map[string]int
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewChatAIService(config *ChatAIConfig) *ChatAIService {
	s := &ChatAIService{
		config:       config,
		userNum:      map[string]int{},
		chatGpt:      gpt.NewService(config.GPT, config.MaxTokens, config.EnableContext),
		ernieBot:     ernie.NewService(config.ErnieBot, config.MaxTokens, config.EnableContext),
		day:          today(),
		limitContent: fmt.Sprintf("You've reached the daily limit (%d/day). Your quota will be restored tomorrow.", config.EverydayLimit),
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	go func() {
		now := time.Now().UTC()
		next := now.Add(time.Hour)
		next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())
		t := time.NewTicker(next.Sub(now))
		select {
		case <-s.ctx.Done():
			t.Stop()
			return
		case <-t.C:
			s.reset()
		}
		t.Stop()
		tk := time.NewTicker(time.Hour)
		defer tk.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-tk.C:
				s.reset()
			}
		}
	}()
	return s
}

func (s *ChatAIService) Close() {
	s.cancel()
}

func (s *ChatAIService) addNum(userIdentifier string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	num, ok := s.userNum[userIdentifier]
	if !ok {
		num = 0
	}
	num++
	s.userNum[userIdentifier] = num
}

func (s *ChatAIService) getNum(userIdentifier string) int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	num, ok := s.userNum[userIdentifier]
	if !ok {
		return 0
	}
	return num
}

func (s *ChatAIService) reset() {
	s.lock.Lock()
	defer s.lock.Unlock()
	time.Sleep(time.Second)
	t := today()
	if s.day == t {
		return
	}
	s.day = t
	s.userNum = map[string]int{}
	s.chatGpt.Reset()
	s.ernieBot.Reset()
}

func (s *ChatAIService) ChatCompletion(chatType Type, userIdentifier, content string) (resp string, err error) {
	if content == "" {
		return "", nil
	}
	if chatType == "" {
		chatType = s.config.DefaultChatMode
	}
	num := s.getNum(userIdentifier)
	if num >= s.config.EverydayLimit {
		return s.limitContent, ErrLimit
	}
	log.Infof("receive content: %s, type: %s, user: %s\n", content, chatType, userIdentifier)
	switch chatType {
	case ChatErnieBot:
		resp, err = s.ernieBot.SendChatCompletion(userIdentifier, content)
	case ChatGPT:
		resp, err = s.chatGpt.SendChatCompletion(userIdentifier, content)
	}
	if err != nil {
		return responseErr, err
	}
	if resp == "" {
		return responseEmpty, nil
	}
	s.addNum(userIdentifier)
	return resp, nil
}

type ChatAIClientInterfacer interface {
	SendChatCompletion(userIdentifier, content string) (string, error)
	Reset()
}
