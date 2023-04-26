package controller

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	log "k8s.io/klog/v2"

	vanusv1alpha1 "github.com/vanus-labs/vanus-connect-runtime/pkg/apis/vanus/v1alpha1"
)

func (c *Controller) enqueueAddConnector(obj interface{}) {
	if !c.filterConnectors(obj.(*vanusv1alpha1.Connector)) {
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	log.Infof("enqueue add connector %s", key)
	c.addConnectorQueue.Add(key)
}

func (c *Controller) enqueueUpdateConnector(old, new interface{}) {

	oldConnector := old.(*vanusv1alpha1.Connector)
	newConnector := new.(*vanusv1alpha1.Connector)
	if !c.filterConnectors(new.(*vanusv1alpha1.Connector)) {
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(new); err != nil {
		utilruntime.HandleError(err)
		return
	}

	log.Infof("old connector: %+v\n", oldConnector)
	log.Infof("new connector: %+v\n", newConnector)
	c.updateConnectorQueue.Add(key)
}

func (c *Controller) enqueueDeleteConnector(obj interface{}) {
	if !c.filterConnectors(obj.(*vanusv1alpha1.Connector)) {
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	log.Infof("enqueue delete connector %s", key)
	c.deleteConnectorQueue.Add(obj)
}

func (c *Controller) runAddConnectorWorker() {
	for c.processNextAddConnectorWorkItem() {
	}
}

func (c *Controller) runUpdateConnectorWorker() {
	for c.processNextUpdateConnectorWorkItem() {
	}
}

func (c *Controller) runDeleteConnectorWorker() {
	for c.processNextDeleteConnectorWorkItem() {
	}
}

func (c *Controller) processNextAddConnectorWorkItem() bool {
	obj, shutdown := c.addConnectorQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.addConnectorQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.addConnectorQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.handleAddConnector(key); err != nil {
			c.addConnectorQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.addConnectorQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) processNextUpdateConnectorWorkItem() bool {
	obj, shutdown := c.updateConnectorQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.updateConnectorQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.updateConnectorQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.handleUpdateConnector(key); err != nil {
			c.updateConnectorQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.updateConnectorQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) processNextDeleteConnectorWorkItem() bool {
	obj, shutdown := c.deleteConnectorQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.deleteConnectorQueue.Done(obj)
		var connector *vanusv1alpha1.Connector
		var ok bool
		if connector, ok = obj.(*vanusv1alpha1.Connector); !ok {
			c.deleteConnectorQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected connector in workqueue but got %#v", obj))
			return nil
		}
		if err := c.handleDeleteConnector(connector); err != nil {
			c.deleteConnectorQueue.AddRateLimited(obj)
			return fmt.Errorf("error syncing '%s': %s, requeuing", connector.Name, err.Error())
		}
		c.deleteConnectorQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) filterConnectors(connector *vanusv1alpha1.Connector) bool {
	return c.filterConnector.Kind == connector.Spec.Kind && c.filterConnector.Type == c.filterConnector.Type
}

func (c *Controller) handleAddConnector(key string) error {
	var err error
	cachedConnector, err := c.connectorsLister.Get(key)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	log.Infof("handle add connector %s", cachedConnector.Name)
	err = c.connectorHandler.OnAdd(cachedConnector.GetConnectorID(), cachedConnector.Spec.Config)
	if err != nil {
		log.Infof("handle add connector %s failed %v", cachedConnector.Name, err)
		return err
	}
	return nil
}

func (c *Controller) handleUpdateConnector(key string) error {
	var err error
	cachedConnector, err := c.connectorsLister.Get(key)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	log.Infof("handle update connector %s", cachedConnector.Name)
	err = c.connectorHandler.OnAdd(cachedConnector.GetConnectorID(), cachedConnector.Spec.Config)
	if err != nil {
		log.Infof("handle update connector %s failed %v", cachedConnector.Name, err)
		return err
	}
	return nil
}

func (c *Controller) handleDeleteConnector(connector *vanusv1alpha1.Connector) error {
	log.Infof("handle delete connector %s", connector.Name)
	err := c.connectorHandler.OnDelete(connector.GetConnectorID())
	if err != nil {
		log.Infof("handle delete connector %s failed %v", connector.Name, err)
		return err
	}
	return nil
}
