package main

import (
	"context"
	"fmt"
	"time"

	"github.com/vanus-labs/vanus-connect-runtime/pkg/controller"
	"k8s.io/apimachinery/pkg/labels"
)

func main() {
	connectorEventHandlerFuncs := controller.ConnectorEventHandlerFuncs{
		AddFunc: func(connectorID, config string) error {
			fmt.Printf("===AddFunc=== %s\n", connectorID)
			return nil
		},
		UpdateFunc: func(connectorID, config string) error {
			fmt.Printf("===UpdateFunc=== %s\n", connectorID)
			return nil
		},
		DeleteFunc: func(connectorID string) error {
			fmt.Printf("===DeleteFunc=== %s\n", connectorID)
			return nil
		},
	}
	ctrl, err := controller.NewController(controller.WithFilter("kind=source,type=chatgpt"), controller.WithEventHandler(connectorEventHandlerFuncs))
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	go ctrl.Run(ctx)
	time.Sleep(3 * time.Second)
	cs, err := ctrl.Lister().List(labels.Everything())
	if err != nil {
		fmt.Println("failed to list connectors")
		panic(err)
	}
	for k, v := range cs {
		fmt.Printf("===list connectors=== k: %+v, v: %+v\n", k, v)
	}
	id := "my-connector"
	connector, err := ctrl.Lister().Get(id)
	if err != nil {
		fmt.Printf("failed to get connector %s\n", id)
		panic(err)
	}
	fmt.Printf("success to get connector %+v\n", connector)
	<-ctx.Done()
}
