package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RemediationType contains type of the remediation
type RemediationType string

const (
	// RemediationTypeReboot contains reboot type of the remediation
	RemediationTypeReboot RemediationType = "Reboot"
	// RemediationTypeRecreate contains re-create type of the remediation
	RemediationTypeRecreate RemediationType = "Re-Create"
)

// RemediationState contains state of the remediation
type RemediationState string

const (
	// RemediationStateStarted contains remediation state when the operation started
	RemediationStateStarted RemediationState = "Started"
	// RemediationStateInProgress contains remediation state when the operation in progress
	RemediationStateInProgress RemediationState = "InProgress"
	// RemediationStateSucceeded contains remediation state when the operation succeeded
	RemediationStateSucceeded RemediationState = "Succeeded"
	// RemediationStateFailed contains remediation state when the operation failed
	RemediationStateFailed RemediationState = "Failed"
	// RemediationStateFailedOnTimeout contains remediation state when the operation failed on the timeout
	RemediationStateFailedOnTimeout RemediationState = "FailedOnTimeout"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineRemediationRequest is the schema for the machineremediationrequest API
// kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mrr;mrrs
// +k8s:openapi-gen=true
type MachineRemediationRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of MachineRemediationRequest
	Spec MachineRemediationRequestSpec `json:"spec,omitempty"`

	// Most recently observed status of MachineRemediationRequest resource
	Status MachineRemediationRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineRemediationRequestList contains a list of MachineRemediationRequest
type MachineRemediationRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineRemediationRequest `json:"items"`
}

// MachineRemediationRequestSpec defines the spec of MachineRemediationRequest
type MachineRemediationRequestSpec struct {
	// Type contains the type of the remediation
	Type RemediationType `json:"type,omitempty" valid:"required"`
	// MachineName contains the name of machine that should be remediate
	MachineName string `json:"machineName,omitempty" valid:"required"`
}

// MachineRemediationRequestStatus defines the observed status of MachineRemediationRequest
type MachineRemediationRequestStatus struct {
	State     RemediationState `json:"state,omitempty"`
	StartTime metav1.Time      `json:"startTime,omitempty"`
}
