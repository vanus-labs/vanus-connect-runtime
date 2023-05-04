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
	"github.com/go-openapi/runtime/middleware"

	"github.com/vanus-labs/vanus-connect-runtime/api/models"
	"github.com/vanus-labs/vanus-connect-runtime/api/restapi/operations/connector"
	log "k8s.io/klog/v2"
)

// All registered processing functions should appear under Registxxx in order
func RegistChatAIHandler(a *Api) {
	a.ConnectorChataiHandler = connector.ChataiHandlerFunc(a.chataiHandler)
}

func (a *Api) chataiHandler(params connector.ChataiParams) middleware.Responder {
	// TODO(jiangkai): implement functions
	log.Infof("show chatai params, connector_id: %s, message: %s\n",
		params.ConnectorID,
		params.Message)
	return connector.NewChatgptOK().WithPayload(&models.APIResponse{
		Code:    200,
		Message: "this is chatgpt answer",
	})
}