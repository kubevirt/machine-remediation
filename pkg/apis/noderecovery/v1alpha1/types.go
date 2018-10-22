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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterapiv1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

const NamespaceNoderecovery = "noderecovery"

const (
	EventNodeRemediationCreateFailed     = "FailedCreate"
	EventNodeRemediationCreateSuccessful = "SuccessfulCreate"
	EventNodeRemediationDeleteFailed     = "FailedDelete"
	EventNodeRemediationDeleteSuccessful = "SuccessfulDelete"
	EventNodeRemediationUpdateFailed     = "FailedUpdate"
	EventNodeRemediationUpdateSuccessful = "SuccessfulUpdate"
	EventNodeRemediationFailed           = "NodeRemediationFailed"
	EventNodeRemediationSucceeded        = "NodeRemediationSucceeded"
)

const (
	EventMachineCreateFailed     = "FailedCreate"
	EventMachineCreateSuccessful = "SuccessfulCreate"
	EventMachineDeleteFailed     = "FailedDelete"
	EventMachineDeleteSuccessful = "SuccessfulDelete"
)

const (
	MachineAPIVersion = "cluster.k8s.io/v1alpha1"
	MachineKind       = "Machine"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeRemediation is a specification for a NodeRemediation resource
type NodeRemediation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *NodeRemediationSpec   `json:"spec,omitempty" valid:"required"`
	Status *NodeRemediationStatus `json:"status,omitempty"`
}

type NodeRemediationSpec struct {
	// NodeName maps between node and NodeRemediation object
	NodeName         string                         `json:"nodeName,omitempty" valid:"required"`
	MachineCluster   string                         `json:"machineCluster,omitempty"`
	MachineName      string							`json:"machineName,omitempty"`
	MachineNamespace string                         `json:"machineNamespace,omitempty"`
	MachineSpec      clusterapiv1alpha1.MachineSpec `json:"machineSpec,omitempty"`
}

type NodeRemediationPhase string

const (
	NodeRemediationPhaseInit      NodeRemediationPhase = "Init"
	NodeRemediationPhaseWait      NodeRemediationPhase = "Wait"
	NodeRemediationPhaseRemediate NodeRemediationPhase = "Remediate"
)

type NodeRemediationStatus struct {
	// Reason gives a brief CamelCase message indicating details about why the NodeRemediation is in this state. e.g. 'Init'
	// +optional
	Reason string `json:"reason,omitempty"`
	// Phase indicates NodeRemediation recovery state
	Phase NodeRemediationPhase `json:"phase,omitempty"`
	// StartTime indicates when current phase started
	StartTime metav1.Time `json:"startTime,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeRemediationList is a list of NodeRemediation resources
type NodeRemediationList struct {
	// Map of node names to recovery efforts
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []NodeRemediation `json:"items"`
}
