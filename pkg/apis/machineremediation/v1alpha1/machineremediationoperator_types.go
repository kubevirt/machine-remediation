package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperatorConditionType is the state of the operator's reconciliation functionality.
type OperatorConditionType string

const (
	// OperatorAvailable indicates that the binary maintained by the operator is functional and available in the cluster.
	OperatorAvailable OperatorConditionType = "Available"

	// OperatorProgressing indicates that the operator is actively making changes to the binary maintained by the operator.
	OperatorProgressing OperatorConditionType = "Progressing"

	// OperatorDegraded indicates that the operand is not functioning completely. An example of a degraded state
	// would be if there should be 5 copies of the operand running but only 4 are running. It may still be available,
	// but it is degraded
	OperatorDegraded OperatorConditionType = "Degraded"
)

// OperatorConditionStatus contains the status of the operator condition.
type OperatorConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    OperatorConditionStatus = "True"
	ConditionFalse   OperatorConditionStatus = "False"
	ConditionUnknown OperatorConditionStatus = "Unknown"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineRemediationOperator is the schema for the MachineRemediationOperator API
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mro;mros
// +k8s:openapi-gen=true
type MachineRemediationOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of MachineRemediationOperator
	Spec MachineRemediationOperatorSpec `json:"spec,omitempty"`

	// Most recently observed status of MachineRemediationOperator resource
	Status MachineRemediationOperatorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineRemediationOperatorList contains a list of MachineRemediationOperator
type MachineRemediationOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineRemediationOperator `json:"items"`
}

// MachineRemediationOperatorSpec defines the spec of MachineRemediation
type MachineRemediationOperatorSpec struct {
	// The image tag to use for the continer images installed.
	// Defaults to the same tag as the operator's container image.
	ImageTag string `json:"imageTag,omitempty"`
	// The image registry to pull the container images from
	// Defaults to the same registry the operator's container image is pulled from.
	ImageRegistry string `json:"imageRegistry,omitempty"`
	// The ImagePullPolicy to use.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" valid:"required"`
}

// MachineRemediationOperatorStatus defines the observed status of MachineRemediationOperator
type MachineRemediationOperatorStatus struct {
	// type specifies the state of the operator's reconciliation functionality,
	// which reflects the state of the application
	Conditions []MachineRemediationOperatorStatusCondition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`
}

// MachineRemediationOperatorStatusCondition represents the state of the operator's
// reconciliation functionality.
// +k8s:deepcopy-gen=true
type MachineRemediationOperatorStatusCondition struct {
	// type specifies the state of the operator's reconciliation functionality,
	// which reflects the state of the application
	Type OperatorConditionType `json:"type"`

	// status of the condition, one either True or False.
	Status OperatorConditionStatus `json:"status"`

	// lastTransitionTime is the time of the last update to the current status object.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// reason is the reason for the condition's last transition.  Reasons are CamelCase
	Reason string `json:"reason,omitempty"`

	// message provides additional information about the current condition.
	// This is only to be consumed by humans.
	Message string `json:"message,omitempty"`
}
