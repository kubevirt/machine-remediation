package components

import (
	maov1 "github.com/openshift/machine-api-operator/pkg/apis/healthchecking/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/machine-remediation-operator/pkg/consts"
)

// NewMastersMachineHealthCheck retruns new MachineHealthCheck object for master nodes
func NewMastersMachineHealthCheck(name string, namespace string, operatorVersion string) *maov1.MachineHealthCheck {
	return &maov1.MachineHealthCheck{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "machine.openshift.io/v1beta1",
			Kind:       "MachineHealthCheck",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				maov1.SchemeGroupVersion.Group:              "",
				maov1.SchemeGroupVersion.Group + "/version": operatorVersion,
			},
		},
		Spec: maov1.MachineHealthCheckSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					consts.MachineRoleLabel: "master",
				},
			},
		},
	}
}
