package controller

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	log "k8s.io/klog/v2"

	vanusv1alpha1 "github.com/vanus-labs/vanus-connect-runtime/pkg/apis/vanus/v1alpha1"
)

func (c *controller) enqueueAddConnector(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	log.Infof("enqueue add connector %s", key)
	c.addConnectorQueue.Add(key)
}

func (c *controller) enqueueUpdateConnector(old, new interface{}) {
	oldKey, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	newKey, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	log.Infof("old connector: %s\n", oldKey)
	log.Infof("new connector: %s\n", newKey)
	c.updateConnectorQueue.Add(newKey)
}

func (c *controller) enqueueDeleteConnector(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	log.Infof("enqueue delete connector %s\n", key)
	c.deleteConnectorQueue.Add(obj)
}

func (c *controller) runAddConnectorWorker() {
	for c.processNextAddConnectorWorkItem() {
	}
}

func (c *controller) runUpdateConnectorWorker() {
	for c.processNextUpdateConnectorWorkItem() {
	}
}

func (c *controller) runDeleteConnectorWorker() {
	for c.processNextDeleteConnectorWorkItem() {
	}
}

func (c *controller) processNextAddConnectorWorkItem() bool {
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

func (c *controller) processNextUpdateConnectorWorkItem() bool {
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

func (c *controller) processNextDeleteConnectorWorkItem() bool {
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

func (c *controller) handleAddConnector(key string) error {
	var err error
	cachedConnector, err := c.connectorsLister.Get(key)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	log.Infof("handle add connector %s", cachedConnector.Name)
	err = c.handler.OnAdd(cachedConnector.Name, cachedConnector.Spec.Config)
	if err != nil {
		log.Errorf("handle add connector %s failed: %+v", cachedConnector.Name, err)
		return err
	}
	return nil
}

func (c *controller) handleUpdateConnector(key string) error {
	var err error
	cachedConnector, err := c.connectorsLister.Get(key)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	log.Infof("handle update connector %s", cachedConnector.Name)
	err = c.handler.OnUpdate(cachedConnector.Name, cachedConnector.Spec.Config)
	if err != nil {
		log.Errorf("handle update connector %s failed: %+v", cachedConnector.Name, err)
		return err
	}
	return nil
}

func (c *controller) handleDeleteConnector(connector *vanusv1alpha1.Connector) error {
	log.Infof("handle delete connector %s", connector.Name)
	err := c.handler.OnDelete(connector.Name)
	if err != nil {
		log.Errorf("handle delete connector %s failed: %+v", connector.Name, err)
		return err
	}
	return nil
}
