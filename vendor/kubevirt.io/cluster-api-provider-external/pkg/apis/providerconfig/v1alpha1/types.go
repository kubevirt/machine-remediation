/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The MachineRole indicates the purpose of the Machine, and will determine
// what software and configuration will be used when provisioning and managing
// the Machine. A single Machine may have more than one role, and the list and
// definitions of supported roles is expected to evolve over time.
//
// Currently, only two roles are supported: Master and Node. In the future, we
// expect user needs to drive the evolution and granularity of these roles,
// with new additions accommodating common cluster patterns, like dedicated
// etcd Machines.
//
//                 +-----------------------+------------------------+
//                 | Master present        | Master absent          |
// +---------------+-----------------------+------------------------|
// | Node present: | Install control plane | Join the cluster as    |
// |               | and be schedulable    | just a node            |
// |---------------+-----------------------+------------------------|
// | Node absent:  | Install control plane | Invalid configuration  |
// |               | and be unschedulable  |                        |
// +---------------+-----------------------+------------------------+

type MachineRole string

const (
	MasterRole MachineRole = "Master"
	NodeRole   MachineRole = "Node"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExternalMachineProviderConfig provides machine configuration struct
type ExternalMachineProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	// FencingConfig specify machine power management configuration
	FencingConfig *FencingConfig `json:"fencingConfig"`
	// Label give possibility to map between machine to specific configuration under configMap
	Label string `json:"label,omitempty"`
	// Roles specify which role will server machine under the cluster
	Roles []MachineRole `json:"roles,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExternalClusterProviderConfig provides machine configuration struct
type ExternalClusterProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	Project        string          `json:"project"`
	FencingConfigs []FencingConfig `json:"fencingConfigs"`
}

type FencingConfig struct {
	metav1.ObjectMeta `json:",inline"`

	// Container that handles machine operations
	Container *v1.Container `json:"container"`

	// Optional command to be used instead of the default when
	// handling machine Create operations (power-on/provisioning)
	CheckArgs []string `json:"checkArgs,omitempty"`

	// Optional command to be used instead of the default when
	// handling machine Create operations (power-on/provisioning)
	CreateArgs []string `json:"createArgs,omitempty"`

	// Optional command to be used instead of the default when
	// handling machine Delete operations (power-off/deprovisioning)
	DeleteArgs []string `json:"deleteArgs,omitempty"`

	// Optional command to be used instead of the default when
	// handling machine Update operations (reboot)
	RebootArgs []string `json:"rebootArgs,omitempty"`

	// Parameters common to all commands that may be passed as either
	// name/value pairs or "--name value" depending on the value of
	// ArgumentFormat
	Config map[string]string `json:"config"`

	// Parameters who’s value changes depending on the affected node
	// Not relevant if setting as part of the machine definition
	DynamicConfig []DynamicConfigElement `json:"dynamicConfig,omitempty"`

	// Secret contains fencing agent username and password
	Secret string `json:"secret"`

	// How long to wait for the Job to complete
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty"`

	// How long to wait before retrying failed Jobs
	RetrySeconds *int32 `json:"retrySeconds,omitempty"`

	// How long to wait before retrying failed Jobs
	Retries *int32 `json:"retries,omitempty"`

	// Volumes represent additional volumes that you want to attach
	// to the container
	Volumes []v1.Volume `json:"volumes,omitempty"`
}

type DynamicConfigElement struct {
	Field   string            `json:"field"`
	Default *string           `json:"default,omitempty"`
	Values  map[string]string `json:"values"`
}

func (dc *DynamicConfigElement) Lookup(key string) (string, bool) {
	if val, ok := dc.Values[key]; ok {
		return val, true
	} else if dc.Default != nil {
		return *dc.Default, true
	}
	return "", false
}
