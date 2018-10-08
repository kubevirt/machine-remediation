/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package noderecovery

import (
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"kubevirt.io/node-recovery/pkg/apis/noderecovery/v1alpha1"
	clientset "kubevirt.io/node-recovery/pkg/client/clientset/versioned"
	informers "kubevirt.io/node-recovery/pkg/client/informers/externalversions/noderecovery/v1alpha1"
	listers "kubevirt.io/node-recovery/pkg/client/listers/noderecovery/v1alpha1"
	"kubevirt.io/node-recovery/pkg/controller"
)

const (
	// maxRetries is the number of times a deployment will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

type NodeRecoveryController struct {
	kubeclientset         kubernetes.Interface
	noderecoveryclientset clientset.Interface

	queue workqueue.RateLimitingInterface

	nodeLister            corelisters.NodeLister
	nodeSynced            cache.InformerSynced
	configMapLister       corelisters.ConfigMapLister
	configMapSynced       cache.InformerSynced
	nodeRemediationLister listers.NodeRemediationLister
	nodeRemediationSynced cache.InformerSynced

	nodeConditionManager *controller.NodeConditionManager
}

// NewNodeRecoveryController returns new NodeRecoveryController instance
func NewNodeRecoveryController(
	kubeclientset kubernetes.Interface,
	noderecoveryclientset clientset.Interface,
	nodeInformer coreinformers.NodeInformer,
	configMapInformer coreinformers.ConfigMapInformer,
	nodeRemediationInformer informers.NodeRemediationInformer,
) *NodeRecoveryController {

	c := &NodeRecoveryController{
		kubeclientset:         kubeclientset,
		noderecoveryclientset: noderecoveryclientset,
		queue:                 workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		nodeLister:            nodeInformer.Lister(),
		nodeSynced:            nodeInformer.Informer().HasSynced,
		configMapLister:       configMapInformer.Lister(),
		configMapSynced:       configMapInformer.Informer().HasSynced,
		nodeRemediationLister: nodeRemediationInformer.Lister(),
		nodeRemediationSynced: nodeRemediationInformer.Informer().HasSynced,
		nodeConditionManager:  controller.NewNodeConditionManager(),
	}

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addNode,
		DeleteFunc: c.deleteNode,
		UpdateFunc: c.updateNode,
	})

	nodeRemediationInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addRemediationNode,
		DeleteFunc: c.deleteRemediationNode,
		UpdateFunc: c.updateRemediationNode,
	})

	c.nodeConditionManager = controller.NewNodeConditionManager()

	return c
}

// TODO: I think we will need only part of this handlers,
// addNode, deleteNode, updateNode and maybe
func (c *NodeRecoveryController) addNode(obj interface{}) {
}

func (c *NodeRecoveryController) deleteNode(obj interface{}) {
}

func (c *NodeRecoveryController) updateNode(old, cur interface{}) {
	curNode := cur.(*apiv1.Node)
	oldNode := old.(*apiv1.Node)
	if curNode.ResourceVersion == oldNode.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}

	if curNode.DeletionTimestamp != nil {
		return
	}

	if !reflect.DeepEqual(curNode.Status, oldNode.Status) {
		glog.V(2).Infof("node %s status updated", curNode.Name)
		c.enqueueNode(cur)
	}
}

func (c *NodeRecoveryController) enqueueNode(obj interface{}) {
	node := obj.(*apiv1.Node)
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(node)
	if err != nil {
		glog.Errorf("failed to extract for the node %s", node.Name)
	}
	c.queue.Add(key)
}

func (c *NodeRecoveryController) addRemediationNode(obj interface{}) {
}

func (c *NodeRecoveryController) deleteRemediationNode(obj interface{}) {
}

func (c *NodeRecoveryController) updateRemediationNode(old, curr interface{}) {
}

// Run begins watching and syncing.
func (c *NodeRecoveryController) Run(threadiness int, stopCh chan struct{}) {
	defer controller.HandlePanic()
	defer c.queue.ShutDown()
	glog.Info("starting node-recovery controller.")

	// Wait for cache sync before we start the pod controller
	if !controller.WaitForCacheSync("node-recovery", stopCh, c.nodeSynced, c.configMapSynced, c.nodeRemediationSynced) {
		return
	}

	// Start the actual work
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("stopping node-recovery controller.")
}

func (c *NodeRecoveryController) worker() {
	for c.processNextWorkItem() {
	}
}

// Execute runs a worker thread that just dequeues items
func (c *NodeRecoveryController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)
	err := c.syncNode(key.(string))

	c.handleErr(err, key)
	return true
}

func (c *NodeRecoveryController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		glog.V(2).Infof("error syncing node %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	glog.V(2).Infof("dropping node %q out of the queue: %v", key, err)
	c.queue.Forget(key)
}

func (c *NodeRecoveryController) syncNode(key string) error {
	// Fetch the latest Vm state from cache
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	node, err := c.nodeLister.Get(name)

	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("node '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	readyCond := c.nodeConditionManager.GetNodeCondition(node, apiv1.NodeReady)

	if readyCond.Status != apiv1.ConditionTrue {
		glog.Infof("node %s has ready condition false", node.Name)
		_, err := c.nodeRemediationLister.NodeRemediations("kube-system").Get(name)
		if err != nil {
			if errors.IsNotFound(err) {
				nodeRemediation := &v1alpha1.NodeRemediation{
					Spec: &v1alpha1.NodeRemediationSpec{
						NodeName: node.Name,
					},
					Status: &v1alpha1.NodeRemediationStatus{
						Phase: v1alpha1.NodeRemediationPhaseInit,
					},
				}
				nodeRemediation.Name = node.Name

				_, err := c.noderecoveryclientset.NoderecoveryV1alpha1().NodeRemediations("kube-system").Create(nodeRemediation)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// TODO: add remediation logic
	return nil
}
