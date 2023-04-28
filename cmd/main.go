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

package main

import (
	"context"
	"flag"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/vanus-labs/vanus-connect-runtime/internal/apiserver/handlers"
	"github.com/vanus-labs/vanus-connect-runtime/pkg/controller"
	log "k8s.io/klog/v2"
)

func main() {
	//address options
	addr := flag.String("addr", ":8080", "listen address")
	basepath := flag.String("baseurl", "/api/v1", "base url prefix")
	configPath := flag.String("config", "./config/runtime.yaml", "the configuration file of runtime")
	log.InitFlags(flag.CommandLine)

	cfg, err := handlers.InitConfig(*configPath)
	if err != nil {
		panic(err)
	}
	cfg.Connectors = &sync.Map{}

	flag.Parse()
	bpath := filepath.Clean(*basepath)
	log.Infof("baseurl is: %v", bpath)
	//api init, include wrap handler
	a, err := handlers.NewApi(bpath, cfg)
	if err != nil {
		panic(err)
	}

	engine := gin.Default()
	engine.NoRoute(func(c *gin.Context) {
		a.Handler().ServeHTTP(c.Writer, c.Request)
	})
	engine.NoMethod(func(c *gin.Context) {
		a.Handler().ServeHTTP(c.Writer, c.Request)
	})

	ctx := context.Background()
	filterConnector := controller.FilterConnector{
		Kind: "source",
		Type: "chatgpt",
	}
	connectorHandlerFuncs := controller.ConnectorHandlerFuncs{
		AddFunc: func(connectorID, config string) error {
			cfg.Connectors.Store(connectorID, config)
			return nil
		},
		UpdateFunc: func(connectorID, config string) error {
			cfg.Connectors.Store(connectorID, config)
			return nil
		},
		DeleteFunc: func(connectorID string) error {
			cfg.Connectors.Delete(connectorID)
			return nil
		},
	}
	c, err := controller.NewController(filterConnector, connectorHandlerFuncs)
	if err != nil {
		log.Errorf("new controller manager failed: %+v\b", err)
		panic(err)
	}
	go c.Run(ctx)

	err = engine.Run(*addr)
	if err != nil {
		panic(err)
	}
}
