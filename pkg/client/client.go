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

	clusterapiclientset "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	noderecoveryclientset "kubevirt.io/node-recovery/pkg/client/clientset/versioned"
)

// GetRESTConfig returns REST client configuration
func GetRESTConfig() *rest.Config {
	// creates the in-cluster config
	config, err := config.GetConfig()
	if err != nil {
		panic(err.Error())
	}

	return config
}

// NewKubeClientSet returns k8s client
func NewKubeClientSet() *kubernetes.Clientset {
	config := GetRESTConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// NewNodeRecoveryClientSet returns node-recovery client
func NewNodeRecoveryClientSet() *noderecoveryclientset.Clientset {
	config := GetRESTConfig()
	// creates the clientset
	clientset, err := noderecoveryclientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// NewClusterAPIClientSet returns cluster-api client
func NewClusterAPIClientSet() *clusterapiclientset.Clientset {
	config := GetRESTConfig()
	// creates the clientset
	clientset, err := clusterapiclientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
