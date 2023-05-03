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
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/go-openapi/runtime/middleware"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	sdkgo "github.com/vanus-labs/sdk/golang"
	"github.com/vanus-labs/vanus-connect-runtime/api/models"
	"github.com/vanus-labs/vanus-connect-runtime/api/restapi/operations/connector"
	"github.com/vanus-labs/vanus-connect-runtime/internal/apiserver/utils"
	"gopkg.in/yaml.v2"
	log "k8s.io/klog/v2"
)

// All registered processing functions should appear under Registxxx in order
func RegistChatGPTHandler(a *Api) {
	a.ConnectorChatgptHandler = connector.ChatgptHandlerFunc(a.chatgptHandler)
}

func (a *Api) chatgptHandler(params connector.ChatgptParams) middleware.Responder {
	c, err := a.ctrl().ControllersLister().Get(NamespaceKey(ConnectorKindSource, ConnectorTypeChatGPT, params.ConnectorID))
	if err != nil {
		log.Warningf("failed to get connector: %s\n", params.ConnectorID)
		return utils.Response(500, err)
	}

	chatgptConfig := &chatGPTConfig{}
	err = yaml.Unmarshal([]byte(c.Spec.Config), chatgptConfig)
	if err != nil {
		return utils.Response(500, err)
	}
	eventSource := params.HTTPRequest.Header.Get(headerSource)
	if eventSource == "" {
		eventSource = defaultEventSource
	}
	eventType := params.HTTPRequest.Header.Get(headerType)
	if eventType == "" {
		eventType = defaultEventType
	}

	// parse config
	u, err := url.Parse(chatgptConfig.Target)
	if err != nil {
		log.Errorf("failed to parse target, connector_id: %s, err: %+v\n", params.ConnectorID, err)
		return utils.Response(500, err)
	}
	if Client == nil {
		// use sdk publish event
		opts := &sdkgo.ClientOptions{
			// TODO(jiangkai): use target host
			Endpoint: "vanus-gateway.vanus:8080",
			Token:    "admin",
		}

		Client, err = sdkgo.Connect(opts)
		if err != nil {
			log.Errorf("failed to connect to Vanus cluster, err: %+v\n", err)
			return utils.Response(500, err)
		}
	}
	subPaths := strings.Split(u.Path, "/")
	p := Client.Publisher(sdkgo.WithEventbus(subPaths[2], subPaths[4]))
	go func(params connector.ChatgptParams) {
		event := ce.NewEvent()
		event.SetID(uuid.New().String())
		event.SetTime(time.Now())
		event.SetType(eventType)
		event.SetSource(eventSource)
		if ChatGPTServer == nil {
			ChatGPTServer = newChatGPTService(&chatGPTConfig{
				Port:          a.config.Port,
				Token:         a.config.OpenAIAPIKey,
				EverydayLimit: 100,
				MaxTokens:     3500,
			})
		}
		content, err := ChatGPTServer.CreateChatCompletion(params.Message)
		if err != nil {
			log.Warningf("failed to get content from ChatGPT: %+v\n", err)
		}

		log.Infof("get content from ChatGPT success, content: %s\n", content)

		event.SetData(ce.ApplicationJSON, map[string]string{
			"content": content,
		})
		err = p.Publish(context.Background(), &event)
		if err != nil {
			log.Errorf("publish event failed, err: %s\n", err.Error())
			return
		}
	}(params)
	return connector.NewChatgptOK().WithPayload(&models.APIResponse{
		Code:    200,
		Message: "success",
	})
}

type chatGPTConfig struct {
	Port          int    `json:"port" yaml:"port"`
	Token         string `json:"token" yaml:"token" validate:"required"`
	Target        string `json:"target" yaml:"target"`
	MaxTokens     int    `json:"max_tokens" yaml:"max_tokens"`
	EverydayLimit int    `json:"everyday_limit" yaml:"everyday_limit"`
}

const (
	defaultEventType   = "vanus-chatGPT-type"
	defaultEventSource = "vanus-chatGPT-source"
	headerSource       = "vanus-source"
	headerType         = "vanus-type"
)

const (
	responseEmpty = "Get response empty."
	responseErr   = "Get response failed."
)

var (
	ChatGPTServer *chatGPTService
	Client        sdkgo.Client
	ErrLimit      = fmt.Errorf("reached the daily limit")
)

type chatGPTService struct {
	client       *openai.Client
	config       *chatGPTConfig
	lock         sync.Mutex
	day          int
	num          int
	limitContent string
}

func newChatGPTService(config *chatGPTConfig) *chatGPTService {
	client := openai.NewClient(config.Token)
	return &chatGPTService{
		config:       config,
		client:       client,
		day:          today(),
		limitContent: fmt.Sprintf("You've reached the daily limit (%d/day). Your quota will be restored tomorrow.", config.EverydayLimit),
	}
}

func today() int {
	return time.Now().UTC().Day()
}

func (s *chatGPTService) reset() {
	s.day = today()
	s.num = 0
}

func (s *chatGPTService) CreateChatCompletion(content string) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.num >= s.config.EverydayLimit {
		if today() == s.day {
			return s.limitContent, ErrLimit
		}
		s.reset()
	}
	log.Infof("receive content: %s\n", content)
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: s.config.MaxTokens,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
	if err != nil {
		return responseErr, err
	}
	s.num++
	if len(resp.Choices) == 0 {
		return responseEmpty, nil
	}
	respContent := resp.Choices[0].Message.Content
	if respContent == "" {
		return responseEmpty, nil
	}
	return respContent, nil
}
