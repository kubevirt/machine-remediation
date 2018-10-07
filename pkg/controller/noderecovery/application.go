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
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"

	"kubevirt.io/node-recovery/pkg/client"
	"kubevirt.io/node-recovery/pkg/controller/informerfactory"
	"kubevirt.io/node-recovery/pkg/controller/leaderelectionconfig"
)

const controllerThreads = 5

type NodeRecoveryImpl struct {
	clientSet client.NodeRecoveryClient

	nodeInformer      cache.SharedIndexInformer
	configMapInformer cache.SharedIndexInformer
	nodeRemediation   cache.SharedIndexInformer
	jobsInformer      cache.SharedIndexInformer

	informerFactory informerfactory.InformerFactory

	leaderElection leaderelectionconfig.Configuration

	controller *NodeRecoveryController
}

func Execute() {
	initializeLogging()

	var nodeRecoveryApp NodeRecoveryImpl
	initializeNodeRecovery(&nodeRecoveryApp)
	nodeRecoveryApp.Run()
}

func initializeNodeRecovery(nodeRecoveryApp *NodeRecoveryImpl) {
	nodeRecoveryApp.clientSet = client.NewNodeRecoveryClient()
	nodeRecoveryApp.informerFactory = informerfactory.NewInformerFactory(nodeRecoveryApp.clientSet)
	nodeRecoveryApp.nodeInformer = nodeRecoveryApp.informerFactory.Node()
	nodeRecoveryApp.configMapInformer = nodeRecoveryApp.informerFactory.ConfigMap()
	nodeRecoveryApp.jobsInformer = nodeRecoveryApp.informerFactory.Job()
	nodeRecoveryApp.leaderElection = leaderelectionconfig.DefaultLeaderElectionConfiguration()

	nodeRecoveryApp.controller = NewNodeRecoveryController(
		nodeRecoveryApp.nodeInformer,
		nodeRecoveryApp.configMapInformer,
		nodeRecoveryApp.jobsInformer,
	)
}

func initializeLogging() {
	flag.Set("logtostderr", "true")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
}

// Run execute
func (nri *NodeRecoveryImpl) Run() {
	stop := make(chan struct{})
	defer close(stop)

	id, err := os.Hostname()
	if err != nil {
		glog.Fatalf("unable to get hostname: %v", err)
	}

	var namespace string
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			namespace = ns
		}
	} else if os.IsNotExist(err) {
		// TODO: Replace leaderelectionconfig.DefaultNamespace with a flag
		namespace = leaderelectionconfig.DefaultNamespace
	} else {
		glog.Fatalf("error searching for namespace in /var/run/secrets/kubernetes.io/serviceaccount/namespace: %v", err)
	}

	// Create new recorder for node-recovery config map
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&clientv1.EventSinkImpl{Interface: clientv1.New(nri.clientSet.CoreV1().RESTClient()).Events(namespace)})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{Component: leaderelectionconfig.DefaultConfigMapName})

	rl, err := resourcelock.New(nri.leaderElection.ResourceLock,
		namespace,
		leaderelectionconfig.DefaultConfigMapName,
		nri.clientSet.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		})
	if err != nil {
		glog.Fatal(err)
	}

	leaderElector, err := leaderelection.NewLeaderElector(
		leaderelection.LeaderElectionConfig{
			Lock:          rl,
			LeaseDuration: nri.leaderElection.LeaseDuration.Duration,
			RenewDeadline: nri.leaderElection.RenewDeadline.Duration,
			RetryPeriod:   nri.leaderElection.RetryPeriod.Duration,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(stopCh <-chan struct{}) {
					nri.informerFactory.Start(stop)
					go nri.controller.Run(controllerThreads, stop)
				},
				OnStoppedLeading: func() {
					glog.Fatal("leaderelection lost")
				},
			},
		})
	if err != nil {
		glog.Fatal(err)
	}

	leaderElector.Run()
	panic("unreachable")
}
