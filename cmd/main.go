package main

import (
	"context"

	"github.com/vanus-labs/vanus-connect-runtime/pkg/controller"
	log "k8s.io/klog/v2"
)

func main() {
	ctx := context.Background()
	c, err := controller.NewController(ctx)
	if err != nil {
		log.Errorf("new controller manager failed: %+v\b", err)
		panic(err)
	}
	go c.Run(ctx)
	<-ctx.Done()
	log.Info("the vanus connect runtime has been shutdown gracefully")
}
