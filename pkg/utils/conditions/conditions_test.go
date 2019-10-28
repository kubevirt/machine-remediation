package conditions

import (
	"reflect"
	"testing"

	mrotesting "kubevirt.io/machine-remediation-operator/pkg/utils/testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	namespace   = "openshift-machine-api"
	correctData = `items:
- name: Ready 
  timeout: 60s
  status: Unknown`
)

func node(name string, ready bool) *v1.Node {
	return mrotesting.NewNode(name, ready, "machineName")
}

func TestGetNodeCondition(t *testing.T) {
	testsCases := []struct {
		node      *corev1.Node
		condition *corev1.NodeCondition
		expected  *corev1.NodeCondition
	}{
		{
			node: node("hasCondition", true),
			condition: &corev1.NodeCondition{
				Type:   corev1.NodeReady,
				Status: corev1.ConditionTrue,
			},
			expected: &corev1.NodeCondition{
				Type:               corev1.NodeReady,
				Status:             corev1.ConditionTrue,
				LastTransitionTime: mrotesting.KnownDate,
			},
		},
		{
			node: node("doesNotHaveCondition", true),
			condition: &corev1.NodeCondition{
				Type:   corev1.NodeMemoryPressure,
				Status: corev1.ConditionTrue,
			},
			expected: nil,
		},
	}

	for _, tc := range testsCases {
		got := GetNodeCondition(tc.node, tc.condition.Type)
		if !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("Test case: %s. Expected: %v, got: %v", tc.node.Name, tc.expected, got)
		}
	}
}
