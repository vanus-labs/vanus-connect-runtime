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

	"github.com/gin-gonic/gin"
	"github.com/vanus-labs/vanus-connect-runtime/internal/apiserver/handlers"
	"github.com/vanus-labs/vanus-connect-runtime/pkg/controller"
	log "k8s.io/klog/v2"
)

func main() {
	//address options
	addr := flag.String("addr", ":8080", "listen address")
	basepath := flag.String("baseurl", "/api/v1", "base url prefix")
	log.InitFlags(flag.CommandLine)

	flag.Parse()
	bpath := filepath.Clean(*basepath)
	log.Infof("baseurl is: %v", bpath)
	//api init, include wrap handler
	a, err := handlers.NewApi(bpath)
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
	c, err := controller.NewController(controller.FilterConnector{}, controller.ConnectorHandlerFuncs{})
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
