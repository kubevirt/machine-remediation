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
 * Copyright 2017, 2018 Red Hat, Inc.
 *
 */

package informerfactory

import (
	"math/rand"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/fields"

	k8sv1 "k8s.io/api/core/v1"
	batchv1 "k8s.io/api/batch/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const systemNamespace = "kube-system"

type newSharedInformer func() cache.SharedIndexInformer

type InformerFactory interface {
	// Starts any informers that have not been started yet
	// This function is thread safe and idempotent
	Start(stopCh <-chan struct{})

	// Watches for nodes
	Node() cache.SharedIndexInformer

	// Watches for ConfigMap objects
	ConfigMap() cache.SharedIndexInformer

	// Watches for Job objects
	Job() cache.SharedIndexInformer
}

type informerFactory struct {
	clientSet kubernetes.Interface
	lock          sync.Mutex
	defaultResync time.Duration

	informers        map[string]cache.SharedIndexInformer
	startedInformers map[string]bool
}

// NewInformerFactory provides factory to create and start new informers
func NewInformerFactory(clientSet kubernetes.Interface) InformerFactory {
	return &informerFactory{
		clientSet: clientSet,
		// Resulting resync period will be between 12 and 24 hours, like the default for k8s
		defaultResync:    resyncPeriod(12 * time.Hour),
		informers:        make(map[string]cache.SharedIndexInformer),
		startedInformers: make(map[string]bool),
	}
}

// Start can be called from multiple controllers in different go routines safely.
// Only informers that have not started are triggered by this function.
// Multiple calls to this function are idempotent.
func (f *informerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for name, informer := range f.informers {
		if f.startedInformers[name] {
			// skip informers that have already started.
			glog.Infof("SKIPPING informer %s", name)
			continue
		}
		glog.Infof("STARTING informer %s", name)
		go informer.Run(stopCh)
		f.startedInformers[name] = true
	}
}

// internal function used to retrieve an already created informer
// or create a new informer if one does not already exist.
// Thread safe
func (f *informerFactory) getInformer(key string, newFunc newSharedInformer) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informer, exists := f.informers[key]
	if exists {
		return informer
	}
	informer = newFunc()
	f.informers[key] = informer

	return informer
}

func (f *informerFactory) Node() cache.SharedIndexInformer {
	return f.getInformer("nodeInformer", func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(f.clientSet.CoreV1().RESTClient(), "nodes", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &k8sv1.Node{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *informerFactory) ConfigMap() cache.SharedIndexInformer {
	// We currently only monitor configmaps in the kube-system namespace
	return f.getInformer("configMapInformer", func() cache.SharedIndexInformer {
		fieldSelector := fields.OneTermEqualSelector("metadata.name", "node-recovery-config")
		lw := cache.NewListWatchFromClient(f.clientSet.CoreV1().RESTClient(), "configmaps", systemNamespace, fieldSelector)
		return cache.NewSharedIndexInformer(lw, &k8sv1.ConfigMap{}, f.defaultResync, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	})
}

func (f *informerFactory) Job() cache.SharedIndexInformer {
	return f.getInformer("jobInformer", func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(f.clientSet.BatchV1().RESTClient(), "jobs", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &batchv1.Job{}, f.defaultResync, cache.Indexers{})
	})
}

// resyncPeriod computes the time interval a shared informer waits before resyncing with the api server
func resyncPeriod(minResyncPeriod time.Duration) time.Duration {
	factor := rand.Float64() + 1
	return time.Duration(float64(minResyncPeriod.Nanoseconds()) * factor)
}
