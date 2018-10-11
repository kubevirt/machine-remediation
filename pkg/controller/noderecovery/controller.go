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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
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

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
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

	glog.V(2).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events(metav1.NamespaceAll)})
	c.recorder = eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "node-recovery-controller"})

	return c
}

// TODO: remove unused handlers
func (c *NodeRecoveryController) addNode(obj interface{}) {
}

func (c *NodeRecoveryController) deleteNode(obj interface{}) {
}

func (c *NodeRecoveryController) updateNode(old, curr interface{}) {
	currNode := curr.(*corev1.Node)
	oldNode := old.(*corev1.Node)
	if currNode.ResourceVersion == oldNode.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}

	if currNode.DeletionTimestamp != nil {
		return
	}

	if !reflect.DeepEqual(currNode.Status, oldNode.Status) {
		glog.V(2).Infof("node %s status updated", currNode.Name)
		c.enqueueNode(curr)
	}
}

func (c *NodeRecoveryController) addRemediationNode(obj interface{}) {
	c.enqueueNode(obj)
}

func (c *NodeRecoveryController) deleteRemediationNode(obj interface{}) {
	// TODO: if node ready, skip enqueue, otherwise enqueue with some limit on number of tries
}

func (c *NodeRecoveryController) updateRemediationNode(old, curr interface{}) {
	currNodeRemediation := curr.(*v1alpha1.NodeRemediation)
	oldNodeRemediation := old.(*v1alpha1.NodeRemediation)
	if currNodeRemediation.ResourceVersion == oldNodeRemediation.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}

	if currNodeRemediation.DeletionTimestamp != nil {
		return
	}
	c.enqueueNode(curr)
}

func (c *NodeRecoveryController) enqueueNode(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		glog.Errorf("failed to extract key for the node")
	}
	c.queue.Add(key)
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

	readyCond := c.nodeConditionManager.GetNodeCondition(node, corev1.NodeReady)
	nodeReady := readyCond.Status == corev1.ConditionTrue

	nodeRemediation, err := c.nodeRemediationLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			if !nodeReady {
				nodeRemediation := &v1alpha1.NodeRemediation{
					Spec: &v1alpha1.NodeRemediationSpec{
						NodeName: node.Name,
					},
					Status: &v1alpha1.NodeRemediationStatus{
						Phase:       v1alpha1.NodeRemediationPhaseInit,
						Reason:      "initilize node remediation",
						StartTime:   metav1.Now(),
						ElapsedTime: 0,
					},
				}
				nodeRemediation.Name = node.Name

				err := c.createNodeRemediationWithEvent(nodeRemediation, node)
				if err != nil {
					return err
				}
			}
			return nil
		}
		return err
	}

	copyNodeRemediation := nodeRemediation.DeepCopy()
	switch nodeRemediation.Status.Phase {
	case v1alpha1.NodeRemediationPhaseInit:
		if !nodeReady {
			copyNodeRemediation.Status.Phase = v1alpha1.NodeRemediationPhaseWait
			copyNodeRemediation.Status.Reason = "wait to be sure that it does not transient error"
			copyNodeRemediation.Status.StartTime = metav1.Now()
			copyNodeRemediation.Status.ElapsedTime = 0
			err := c.updateNodeRemediationWithEvent(nodeRemediation, copyNodeRemediation, node)
			if err != nil {
				return err
			}
			return nil
		}

		err := c.deleteNodeRemediationWithEvent(nodeRemediation, node)
		if err != nil {
			return err
		}
		return nil
	case v1alpha1.NodeRemediationPhaseWait:
		if !nodeReady {
			currentTime := metav1.Now()
			// TODO: get timeout from configMap
			if copyNodeRemediation.Status.StartTime.Add(time.Minute).After(currentTime.Time) {
				copyNodeRemediation.Status.ElapsedTime = time.Since(copyNodeRemediation.Status.StartTime.Time)
			} else {
				copyNodeRemediation.Status.Phase = v1alpha1.NodeRemediationPhaseRemediate
				copyNodeRemediation.Status.Reason = "remediate the node"
				copyNodeRemediation.Status.StartTime = currentTime
				copyNodeRemediation.Status.ElapsedTime = 0
			}

			err := c.updateNodeRemediationWithEvent(nodeRemediation, copyNodeRemediation, node)
			if err != nil {
				return err
			}
			// TODO: call to machine interface to create and delete the relevant node object
			return nil
		}

		err := c.deleteNodeRemediationWithEvent(nodeRemediation, node)
		if err != nil {
			return err
		}
		return nil
	case v1alpha1.NodeRemediationPhaseRemediate:
		if !nodeReady {
			currentTime := metav1.Now()
			// TODO: get timeout from configMap or the node label(annotation)
			if copyNodeRemediation.Status.StartTime.Add(5 * time.Minute).After(currentTime.Time) {
				copyNodeRemediation.Status.ElapsedTime = time.Since(copyNodeRemediation.Status.StartTime.Time)
				err := c.updateNodeRemediationWithEvent(nodeRemediation, copyNodeRemediation, node)
				if err != nil {
					return err
				}
				return nil
			}
			c.recorder.Eventf(
				node,
				corev1.EventTypeWarning,
				v1alpha1.EventNodeRemediationFailed,
				"Failed to remediate the node",
			)
		} else {
			c.recorder.Eventf(
				node,
				corev1.EventTypeNormal,
				v1alpha1.EventNodeRemediationSucceeded,
				"Succeeded to remediate the node",
			)
		}

		err := c.deleteNodeRemediationWithEvent(nodeRemediation, node)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *NodeRecoveryController) createNodeRemediationWithEvent(nodeRemediation *v1alpha1.NodeRemediation, node *corev1.Node) error {
	_, err := c.noderecoveryclientset.NoderecoveryV1alpha1().NodeRemediations().Create(nodeRemediation)
	if err != nil {
		c.recorder.Eventf(
			node,
			corev1.EventTypeWarning,
			v1alpha1.EventNodeRemediationCreateFailed,
			"Failed to create NodeRemediation: %v", err,
		)
		return err
	}
	c.recorder.Eventf(
		node,
		corev1.EventTypeNormal,
		v1alpha1.EventNodeRemediationCreateSuccessful,
		"Succeeded to create NodeRemediation",
	)
	return nil
}

func (c *NodeRecoveryController) deleteNodeRemediationWithEvent(nodeRemediation *v1alpha1.NodeRemediation, node *corev1.Node) error {
	err := c.noderecoveryclientset.NoderecoveryV1alpha1().NodeRemediations().Delete(nodeRemediation.Name, &metav1.DeleteOptions{})
	if err != nil {
		c.recorder.Eventf(
			node,
			corev1.EventTypeWarning,
			v1alpha1.EventNodeRemediationDeleteFailed,
			"Failed to delete NodeRemediation: %v", err,
		)
		return err
	}
	c.recorder.Eventf(
		node,
		corev1.EventTypeNormal,
		v1alpha1.EventNodeRemediationDeleteSuccessful,
		"Succeeded to delete NodeRemediation",
	)
	return nil
}

func (c *NodeRecoveryController) updateNodeRemediationWithEvent(oldNodeRemediation *v1alpha1.NodeRemediation, newNodeRemediation *v1alpha1.NodeRemediation, node *corev1.Node) error {
	_, err := c.noderecoveryclientset.NoderecoveryV1alpha1().NodeRemediations().Update(newNodeRemediation)
	if err != nil {
		if oldNodeRemediation.Status.Phase != newNodeRemediation.Status.Phase {
			c.recorder.Eventf(
				node,
				corev1.EventTypeWarning,
				v1alpha1.EventNodeRemediationUpdateFailed,
				"Failed to update NodeRemediation phase to %s: %v", newNodeRemediation.Status.Phase, err,
			)
		}
		return err
	}
	if oldNodeRemediation.Status.Phase != newNodeRemediation.Status.Phase {
		c.recorder.Eventf(
			node,
			corev1.EventTypeNormal,
			v1alpha1.EventNodeRemediationUpdateSuccessful,
			"Succeeded to update NodeRemediation phase to %s", newNodeRemediation.Status.Phase,
		)
	}
	return nil
}
