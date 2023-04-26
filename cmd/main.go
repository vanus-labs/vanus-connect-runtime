package main

import (
	"context"

	log "k8s.io/klog/v2"

	"github.com/vanus-labs/vanus-connect-runtime/pkg/controller"
)

func main() {
	ctx := context.Background()
	c, err := controller.NewController(controller.FilterConnector{}, controller.ConnectorHandlerFuncs{})
	if err != nil {
		log.Errorf("new controller manager failed: %+v\b", err)
		panic(err)
	}
	go c.Run(ctx)
	<-ctx.Done()
	log.Info("the vanus connect runtime has been shutdown gracefully")
}
