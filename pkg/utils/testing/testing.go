package testing

import (
	"fmt"
	"time"

	bmov1 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	mrv1 "github.com/openshift/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	"github.com/openshift/machine-remediation-operator/pkg/consts"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
)

const (
	// NamespaceTest contains the name of the testing namespace
	NamespaceTest = "test"
)

var (
	// KnownDate contains date that can be used under tests
	KnownDate = metav1.Time{Time: time.Date(1985, 06, 03, 0, 0, 0, 0, time.Local)}
)

// NewBareMetalHost returns new bare metal host object that can be used for testing
func NewBareMetalHost(name string, online bool, powerOn bool) *bmov1.BareMetalHost {
	return &bmov1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{Kind: "BareMetalHost"},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   NamespaceTest,
		},
		Spec: bmov1.BareMetalHostSpec{
			Online: online,
		},
		Status: bmov1.BareMetalHostStatus{
			PoweredOn: powerOn,
		},
	}
}

// NewMachine returns new machine object that can be used for testing
func NewMachine(name string, nodeName string, bareMetalHostName string) *mapiv1.Machine {
	return &mapiv1.Machine{
		TypeMeta: metav1.TypeMeta{Kind: "Machine"},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				consts.AnnotationBareMetalHost: fmt.Sprintf("%s/%s", NamespaceTest, bareMetalHostName),
			},
			Name:            name,
			Namespace:       NamespaceTest,
			OwnerReferences: []metav1.OwnerReference{{Kind: "MachineSet"}},
		},
		Spec: mapiv1.MachineSpec{},
		Status: mapiv1.MachineStatus{
			NodeRef: &corev1.ObjectReference{
				Name:      nodeName,
				Namespace: metav1.NamespaceNone,
			},
		},
	}
}

// NewMachineRemediation returns new machine remediation object that can be used for testing
func NewMachineRemediation(name string, machineName string, remediationType mrv1.RemediationType, remediationState mrv1.RemediationState) *mrv1.MachineRemediation {
	return &mrv1.MachineRemediation{
		TypeMeta: metav1.TypeMeta{Kind: "MachineRemediation"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: NamespaceTest,
		},
		Spec: mrv1.MachineRemediationSpec{
			MachineName: machineName,
			Type:        remediationType,
		},
		Status: mrv1.MachineRemediationStatus{
			StartTime: &metav1.Time{Time: time.Now()},
			State:     remediationState,
		},
	}
}

// NewNode returns new node object that can be used for testing
func NewNode(name string, ready bool, machineName string) *corev1.Node {
	nodeReadyStatus := corev1.ConditionTrue
	if !ready {
		nodeReadyStatus = corev1.ConditionUnknown
	}

	return &corev1.Node{
		TypeMeta: metav1.TypeMeta{Kind: "Node"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceNone,
			Annotations: map[string]string{
				consts.AnnotationMachine: fmt.Sprintf("%s/%s", NamespaceTest, machineName),
			},
			Labels: map[string]string{},
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{
					Type:               corev1.NodeReady,
					Status:             nodeReadyStatus,
					LastTransitionTime: KnownDate,
				},
			},
		},
	}
}
