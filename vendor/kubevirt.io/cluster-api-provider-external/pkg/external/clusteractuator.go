/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package external

import (
	"github.com/golang/glog"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterclient "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"

	"kubevirt.io/cluster-api-provider-external/pkg/apis/providerconfig/v1alpha1"
)

type ExternalClusterActuator struct {
	clusterClient       clusterclient.Interface
	providerConfigCodec *v1alpha1.ExternalProviderConfigCodec
}

func NewClusterActuator(clusterclient clusterclient.Interface) (*ExternalClusterActuator, error) {
	codec, err := v1alpha1.NewCodec()
	if err != nil {
		return nil, err
	}

	return &ExternalClusterActuator{
		clusterClient:       clusterclient,
		providerConfigCodec: codec,
	}, nil
}

func (a *ExternalClusterActuator) Reconcile(cluster *clusterv1.Cluster) error {
	glog.Infof("NOT IMPLEMENTED: reconciling cluster %s", cluster.Name)
	return nil
}

func (a *ExternalClusterActuator) Delete(cluster *clusterv1.Cluster) error {
	glog.Infof("NOT IMPLEMENTED: deleting cluster %v.", cluster.Name)
	return nil
}

func (a *ExternalClusterActuator) clusterproviderconfig(providerConfig clusterv1.ProviderConfig) (*v1alpha1.ExternalClusterProviderConfig, error) {
	var config v1alpha1.ExternalClusterProviderConfig
	err := a.providerConfigCodec.DecodeFromProviderConfig(providerConfig, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
