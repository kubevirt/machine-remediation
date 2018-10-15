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
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"

	clusterapiclientset "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	clusterapiinformers "sigs.k8s.io/cluster-api/pkg/client/informers_generated/externalversions"

	"kubevirt.io/node-recovery/pkg/client"
	clientset "kubevirt.io/node-recovery/pkg/client/clientset/versioned"
	informers "kubevirt.io/node-recovery/pkg/client/informers/externalversions"
	"kubevirt.io/node-recovery/pkg/controller"
	"kubevirt.io/node-recovery/pkg/controller/leaderelectionconfig"
)

const controllerThreads = 5

type NodeRecoveryImpl struct {
	kubeclient         kubernetes.Interface
	noderecoveryclient clientset.Interface
	clusterapiclient   clusterapiclientset.Interface

	kubeInformerFactory         kubeinformers.SharedInformerFactory
	nodeRecoveryInformerFactory informers.SharedInformerFactory
	clusterapiInformerFactory   clusterapiinformers.SharedInformerFactory

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
	nodeRecoveryApp.kubeclient = client.NewKubeClientSet()
	nodeRecoveryApp.noderecoveryclient = client.NewNodeRecoveryClientSet()
	nodeRecoveryApp.clusterapiclient = client.NewClusterAPIClientSet()
	nodeRecoveryApp.kubeInformerFactory = kubeinformers.NewSharedInformerFactory(nodeRecoveryApp.kubeclient, controller.DefaultResyncPeriod())
	nodeRecoveryApp.nodeRecoveryInformerFactory = informers.NewSharedInformerFactory(nodeRecoveryApp.noderecoveryclient, controller.DefaultResyncPeriod())
	nodeRecoveryApp.clusterapiInformerFactory = clusterapiinformers.NewSharedInformerFactory(nodeRecoveryApp.clusterapiclient, controller.DefaultResyncPeriod())
	nodeRecoveryApp.leaderElection = leaderelectionconfig.DefaultLeaderElectionConfiguration()

	nodeRecoveryApp.controller = NewNodeRecoveryController(
		nodeRecoveryApp.kubeclient,
		nodeRecoveryApp.noderecoveryclient,
		nodeRecoveryApp.clusterapiclient,
		nodeRecoveryApp.kubeInformerFactory.Core().V1().Nodes(),
		nodeRecoveryApp.kubeInformerFactory.Core().V1().ConfigMaps(),
		nodeRecoveryApp.nodeRecoveryInformerFactory.Noderecovery().V1alpha1().NodeRemediations(),
		nodeRecoveryApp.clusterapiInformerFactory.Cluster().V1alpha1().Machines(),
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

	// Create new event recorder for node-recovery leader election
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: typedcorev1.New(nri.kubeclient.CoreV1().RESTClient()).Events(namespace)})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{Component: leaderelectionconfig.DefaultConfigMapName})

	rl, err := resourcelock.New(nri.leaderElection.ResourceLock,
		namespace,
		leaderelectionconfig.DefaultConfigMapName,
		nri.kubeclient.CoreV1(),
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
					go nri.kubeInformerFactory.Start(stop)
					go nri.nodeRecoveryInformerFactory.Start(stop)
					go nri.clusterapiInformerFactory.Start(stop)
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
