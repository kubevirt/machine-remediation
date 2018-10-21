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
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	clusterapiv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterapiclientset "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	clusterapiinformers "sigs.k8s.io/cluster-api/pkg/client/informers_generated/externalversions/cluster/v1alpha1"
	clusterapilisters "sigs.k8s.io/cluster-api/pkg/client/listers_generated/cluster/v1alpha1"

	"kubevirt.io/node-recovery/pkg/apis/noderecovery/v1alpha1"
	clientset "kubevirt.io/node-recovery/pkg/client/clientset/versioned"
	informers "kubevirt.io/node-recovery/pkg/client/informers/externalversions/noderecovery/v1alpha1"
	listers "kubevirt.io/node-recovery/pkg/client/listers/noderecovery/v1alpha1"
	"kubevirt.io/node-recovery/pkg/controller"
	"kubevirt.io/node-recovery/pkg/client/clientset/versioned/scheme"
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
	kubeclient         kubernetes.Interface
	noderecoveryclient clientset.Interface
	clusterapiclient   clusterapiclientset.Interface

	queue workqueue.RateLimitingInterface

	nodeLister            corelisters.NodeLister
	nodeSynced            cache.InformerSynced
	configMapLister       corelisters.ConfigMapLister
	configMapSynced       cache.InformerSynced
	nodeRemediationLister listers.NodeRemediationLister
	nodeRemediationSynced cache.InformerSynced
	machineLister         clusterapilisters.MachineLister
	machineSynced         cache.InformerSynced

	// nodeConditionManager provides interface to work with node conditions
	nodeConditionManager *controller.NodeConditionManager

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// machineExpectations helps to track Machine creation/deletion by controller
	machineExpectations *controller.UIDTrackingControllerExpectations
	// nodeRemediationExpectations helps to track NodeRemediation creation/deletion by controller
	nodeRemediationExpectations *controller.UIDTrackingControllerExpectations
}

// NewNodeRecoveryController returns new NodeRecoveryController instance
func NewNodeRecoveryController(
	kubeclient kubernetes.Interface,
	noderecoveryclient clientset.Interface,
	clusterapiclient clusterapiclientset.Interface,
	nodeInformer coreinformers.NodeInformer,
	configMapInformer coreinformers.ConfigMapInformer,
	nodeRemediationInformer informers.NodeRemediationInformer,
	machineInformer clusterapiinformers.MachineInformer,
) *NodeRecoveryController {

	c := &NodeRecoveryController{
		kubeclient:                  kubeclient,
		noderecoveryclient:          noderecoveryclient,
		clusterapiclient:            clusterapiclient,
		queue:                       workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		nodeLister:                  nodeInformer.Lister(),
		nodeSynced:                  nodeInformer.Informer().HasSynced,
		configMapLister:             configMapInformer.Lister(),
		configMapSynced:             configMapInformer.Informer().HasSynced,
		nodeRemediationLister:       nodeRemediationInformer.Lister(),
		nodeRemediationSynced:       nodeRemediationInformer.Informer().HasSynced,
		machineLister:               machineInformer.Lister(),
		machineSynced:               machineInformer.Informer().HasSynced,
		nodeConditionManager:        controller.NewNodeConditionManager(),
		machineExpectations:         controller.NewUIDTrackingControllerExpectations(controller.NewControllerExpectations()),
		nodeRemediationExpectations: controller.NewUIDTrackingControllerExpectations(controller.NewControllerExpectations()),
	}

	machineInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addMachine,
		DeleteFunc: c.deleteMachine,
	})

	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.updateNode,
	})

	nodeRemediationInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addNodeRemediation,
		DeleteFunc: c.deleteNodeRemediation,
		UpdateFunc: c.updateNodeRemediation,
	})

	c.nodeConditionManager = controller.NewNodeConditionManager()

	scheme.AddToScheme(kubescheme.Scheme)
	glog.V(2).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclient.CoreV1().Events(metav1.NamespaceAll)})
	c.recorder = eventBroadcaster.NewRecorder(kubescheme.Scheme, corev1.EventSource{Component: "node-recovery-controller"})

	return c
}

func (c *NodeRecoveryController) addMachine(obj interface{}) {
	machine := obj.(*clusterapiv1alpha1.Machine)
	if machine.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.deleteMachine(machine)
		return
	}

	nodeKey, err := c.getNodeKey(machine)
	if err != nil {
		glog.Errorf("failed to get node key: %v", err)
		return
	}
	c.machineExpectations.CreationObserved(nodeKey)
	c.enqueueObj(obj)
}

func (c *NodeRecoveryController) deleteMachine(obj interface{}) {
	machine, ok := obj.(*clusterapiv1alpha1.Machine)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the ReplicaSet
	// changed labels the new deployment will not be woken up till the periodic resync.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		machine, ok = tombstone.Obj.(*clusterapiv1alpha1.Machine)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a Machine %#v", obj))
			return
		}
	}

	nodeKey, err := c.getNodeKey(machine)
	if err != nil {
		glog.Errorf("failed to get node key: %v", err)
		return
	}
	machineKey, err := controller.KeyFunc(machine)
	if err != nil {
		return
	}
	c.machineExpectations.DeletionObserved(nodeKey, machineKey)
	c.enqueueObj(obj)
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

	// TODO: check that node does not master, cluster API does not support remediation
	// of the master node
	if !reflect.DeepEqual(currNode.Status, oldNode.Status) {
		glog.V(2).Infof("node %s status updated", currNode.Name)
		c.enqueueObj(curr)
	}
}

func (c *NodeRecoveryController) addNodeRemediation(obj interface{}) {
	nodeRemediation := obj.(*v1alpha1.NodeRemediation)
	if nodeRemediation.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.deleteNodeRemediation(nodeRemediation)
		return
	}

	key, err := controller.KeyFunc(obj)
	if err != nil {
		return
	}
	c.nodeRemediationExpectations.CreationObserved(key)
	c.enqueueObj(obj)
}

func (c *NodeRecoveryController) deleteNodeRemediation(obj interface{}) {
	_, ok := obj.(*v1alpha1.NodeRemediation)
	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale. If the ReplicaSet
	// changed labels the new deployment will not be woken up till the periodic resync.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		_, ok = tombstone.Obj.(*v1alpha1.NodeRemediation)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a NodeRemediation %#v", obj))
			return
		}
	}
	key, err := controller.KeyFunc(obj)
	if err != nil {
		return
	}
	// TODO: if node ready, skip enqueue, otherwise enqueue with some limit on number of tries
	c.nodeRemediationExpectations.DeletionObserved(key, key)
	c.enqueueObj(obj)
}

func (c *NodeRecoveryController) updateNodeRemediation(old, curr interface{}) {
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
	c.enqueueObj(curr)
}

func (c *NodeRecoveryController) enqueueObj(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err != nil {
		glog.Errorf("failed to extract key for the object")
	}
	c.queue.Add(key)
}

// Run begins watching and syncing.
func (c *NodeRecoveryController) Run(threadiness int, stopCh chan struct{}) {
	defer controller.HandlePanic()
	defer c.queue.ShutDown()
	glog.Info("starting node-recovery controller.")

	// Wait for cache sync before we start the pod controller
	if !controller.WaitForCacheSync("node-recovery", stopCh, c.nodeSynced, c.configMapSynced, c.nodeRemediationSynced, c.machineSynced) {
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
	err := c.sync(key.(string))

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

func (c *NodeRecoveryController) sync(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// TODO: I do not sure if cluster API actuator will not delete the node
	// Fetch the latest node state from cache
	node, err := c.nodeLister.Get(name)

	if err != nil {
		if errors.IsNotFound(err) {
			c.nodeRemediationExpectations.DeleteExpectations(key)
			c.machineExpectations.DeleteExpectations(key)
			runtime.HandleError(fmt.Errorf("node '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if !(c.nodeRemediationExpectations.SatisfiedExpectations(key) && c.machineExpectations.SatisfiedExpectations(key)) {
		return nil
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
						Phase:     v1alpha1.NodeRemediationPhaseInit,
						Reason:    "initilize node remediation",
						StartTime: metav1.Now(),
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
				c.enqueueObj(node)
				return nil
			}
			machine, err := c.getMachine(node)
			if err != nil {
				if errors.IsNotFound(err) {
					glog.V(2).Infof("can not find machine for the node %s", node.Name)
					return nil
				}
				return err
			}

			copyNodeRemediation.Spec.MachineCluster = machine.ClusterName
			copyNodeRemediation.Spec.MachineName = machine.Name
			copyNodeRemediation.Spec.MachineNamespace = machine.Namespace
			machine.Spec.DeepCopyInto(&copyNodeRemediation.Spec.MachineSpec)
			copyNodeRemediation.Status.Phase = v1alpha1.NodeRemediationPhaseRemediate
			copyNodeRemediation.Status.Reason = "remediate the node"
			copyNodeRemediation.Status.StartTime = currentTime

			machineKey, err := controller.KeyFunc(machine)
			if err != nil {
				return err
			}

			c.machineExpectations.ExpectDeletions(key, []string{machineKey})
			err = c.clusterapiclient.ClusterV1alpha1().Machines(machine.Namespace).Delete(machine.Name, &metav1.DeleteOptions{})
			if err != nil {
				c.recorder.Eventf(
					copyNodeRemediation,
					corev1.EventTypeWarning,
					v1alpha1.EventMachineDeleteFailed,
					"Failed to delete machine object",
				)
				c.machineExpectations.DeletionObserved(key, machineKey)
				return err
			}
			c.recorder.Eventf(
				copyNodeRemediation,
				corev1.EventTypeNormal,
				v1alpha1.EventMachineDeleteSuccessful,
				"Succeeded to delete machine object",
			)
			err = c.updateNodeRemediationWithEvent(nodeRemediation, copyNodeRemediation, node)
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
	case v1alpha1.NodeRemediationPhaseRemediate:
		_, err := c.getMachine(node)
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			machine := &clusterapiv1alpha1.Machine{}
			machine.APIVersion = v1alpha1.MachineAPIVersion
			machine.ClusterName = nodeRemediation.Spec.MachineCluster
			machine.Kind = v1alpha1.MachineKind
			machine.Name = nodeRemediation.Spec.MachineName
			machine.Namespace = nodeRemediation.Spec.MachineNamespace
			nodeRemediation.Spec.MachineSpec.DeepCopyInto(&machine.Spec)

			c.machineExpectations.ExpectCreations(key, 1)
			_, err = c.clusterapiclient.ClusterV1alpha1().Machines(machine.Namespace).Create(machine)
			if err != nil {
				c.machineExpectations.CreationObserved(key)
				c.recorder.Eventf(
					copyNodeRemediation,
					corev1.EventTypeWarning,
					v1alpha1.EventMachineCreateFailed,
					"Failed to create machine object",
				)
				return err
			}
			c.recorder.Eventf(
				copyNodeRemediation,
				corev1.EventTypeNormal,
				v1alpha1.EventMachineCreateSuccessful,
				"Succeeded to create machine object",
			)
		}
		if !nodeReady {
			currentTime := metav1.Now()
			// TODO: get timeout from configMap or the node label(annotation)
			if copyNodeRemediation.Status.StartTime.Add(5 * time.Minute).After(currentTime.Time) {
				c.enqueueObj(node)
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

		err = c.deleteNodeRemediationWithEvent(nodeRemediation, node)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *NodeRecoveryController) createNodeRemediationWithEvent(nodeRemediation *v1alpha1.NodeRemediation, node *corev1.Node) error {
	key, err := controller.KeyFunc(nodeRemediation)
	if err != nil {
		return err
	}
	c.nodeRemediationExpectations.ExpectCreations(key, 1)

	_, err = c.noderecoveryclient.NoderecoveryV1alpha1().NodeRemediations().Create(nodeRemediation)
	if err != nil {
		c.nodeRemediationExpectations.CreationObserved(key)
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
	key, err := controller.KeyFunc(nodeRemediation)
	if err != nil {
		return err
	}
	c.nodeRemediationExpectations.ExpectDeletions(key, []string{key})

	err = c.noderecoveryclient.NoderecoveryV1alpha1().NodeRemediations().Delete(nodeRemediation.Name, &metav1.DeleteOptions{})
	if err != nil {
		c.nodeRemediationExpectations.DeletionObserved(key, key)
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
	_, err := c.noderecoveryclient.NoderecoveryV1alpha1().NodeRemediations().Update(newNodeRemediation)
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

func(c *NodeRecoveryController) getNodeKey(machine *clusterapiv1alpha1.Machine) (string, error) {
	predicate := func (node *corev1.Node) bool {
		val, ok := node.Annotations["machine"]
		if !ok {
			return false
		}
		namespace, name, err := cache.SplitMetaNamespaceKey(val)
		if err != nil {
			return false
		}
		if name != machine.Name || namespace != machine.Namespace {
			return false
		}
		return true
	}

	nodes, err := c.nodeLister.ListWithPredicate(predicate)
	if err != nil {
		return "", err
	}
	if len(nodes) < 0 {
		return "", fmt.Errorf("failed to find node that has machine annotation")
	}
	return nodes[0].Name, nil
}

func(c *NodeRecoveryController) getMachine(node *corev1.Node) (*clusterapiv1alpha1.Machine, error) {
	val, ok := node.Annotations["machine"]
	if !ok {
		return nil, fmt.Errorf("failed to get machine aanotation from the node")
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(val)
	if err != nil {
		return nil, err
	}

	machine, err := c.machineLister.Machines(namespace).Get(name)
	if err != nil {
		return nil, err
	}

	return machine, nil
}
