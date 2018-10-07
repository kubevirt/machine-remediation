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

package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	noderecoveryclientset "kubevirt.io/node-recovery/pkg/client/clientset/versioned"
	noderecoveryv1alpha1 "kubevirt.io/node-recovery/pkg/client/clientset/versioned/typed/noderecovery/v1alpha1"
)

func getRESTConfig() *rest.Config {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	return config
}

// NewClientSet returns k8s client
func newClientSet() *kubernetes.Clientset {
	config := getRESTConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// NewNodeRecoveryClientSet returns node-recovery client
func newNodeRecoveryClientSet() *noderecoveryclientset.Clientset {
	config := getRESTConfig()
	// creates the clientset
	clientset, err := noderecoveryclientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

type NodeRecoveryClient interface {
	kubernetes.Interface
	NodeRemediation(namespace string) noderecoveryv1alpha1.NodeRemediationInterface
}

type nodeRecoveryClientImpl struct {
	* kubernetes.Clientset
	noderecoveryClient *noderecoveryclientset.Clientset
}

func (n *nodeRecoveryClientImpl) NodeRemediation(namespace string) noderecoveryv1alpha1.NodeRemediationInterface {
	return n.noderecoveryClient.NoderecoveryV1alpha1().NodeRemediations(namespace)
}

func NewNodeRecoveryClient() *nodeRecoveryClientImpl {
	return &nodeRecoveryClientImpl{
		newClientSet(),
		newNodeRecoveryClientSet(),
	}
}
