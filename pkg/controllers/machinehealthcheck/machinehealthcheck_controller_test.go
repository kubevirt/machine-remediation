package machinehealthcheck

import (
	"context"
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"

	mrv1 "kubevirt.io/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	"kubevirt.io/machine-remediation-operator/pkg/consts"
	"kubevirt.io/machine-remediation-operator/pkg/utils/conditions"
	mrotesting "kubevirt.io/machine-remediation-operator/pkg/utils/testing"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	badConditionsData = `items:
- name: Ready 
  timeout: 60s
  status: Unknown`
)

func init() {
	// Add types to scheme
	mapiv1.AddToScheme(scheme.Scheme)
	mrv1.AddToScheme(scheme.Scheme)
}

func TestHasMatchingLabels(t *testing.T) {
	machine := mrotesting.NewMachine("machine", "node", "")
	testsCases := []struct {
		machine            *mapiv1.Machine
		machineHealthCheck *mrv1.MachineHealthCheck
		expected           bool
	}{
		{
			machine:            machine,
			machineHealthCheck: mrotesting.NewMachineHealthCheck("foobar"),
			expected:           true,
		},
		{
			machine: machine,
			machineHealthCheck: &mrv1.MachineHealthCheck{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "NoMatchingLabels",
					Namespace: consts.NamespaceOpenshiftMachineAPI,
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "MachineHealthCheck",
				},
				Spec: mrv1.MachineHealthCheckSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"no": "match",
						},
					},
				},
				Status: mrv1.MachineHealthCheckStatus{},
			},
			expected: false,
		},
	}

	for _, tc := range testsCases {
		if got := hasMatchingLabels(tc.machineHealthCheck, tc.machine); got != tc.expected {
			t.Errorf("Test case: %s. Expected: %t, got: %t", tc.machineHealthCheck.Name, tc.expected, got)
		}
	}
}

// newFakeReconciler returns a new reconcile.Reconciler with a fake client
func newFakeReconciler(initObjects ...runtime.Object) *ReconcileMachineHealthCheck {
	fakeClient := fake.NewFakeClient(initObjects...)
	return &ReconcileMachineHealthCheck{
		client:    fakeClient,
		namespace: consts.NamespaceOpenshiftMachineAPI,
		recorder:  record.NewFakeRecorder(10),
	}
}

type expectedReconcile struct {
	result reconcile.Result
	error  bool
}

func testReconcile(t *testing.T, remediationWaitTime time.Duration, initObjects ...runtime.Object) {
	// healthy node
	nodeHealthy := mrotesting.NewNode("healthy", true, "machineWithNodehealthy")
	machineWithNodeHealthy := mrotesting.NewMachine("machineWithNodehealthy", nodeHealthy.Name, "")

	// recently unhealthy node
	nodeRecentlyUnhealthy := mrotesting.NewNode("recentlyUnhealthy", false, "machineWithNodeRecentlyUnhealthy")
	nodeRecentlyUnhealthy.Status.Conditions[0].LastTransitionTime = metav1.Time{Time: time.Now()}
	machineWithNodeRecentlyUnhealthy := mrotesting.NewMachine("machineWithNodeRecentlyUnhealthy", nodeRecentlyUnhealthy.Name, "")

	// node without machine annotation
	nodeWithoutMachineAnnotation := mrotesting.NewNode("withoutMachineAnnotation", true, "")
	nodeWithoutMachineAnnotation.Annotations = map[string]string{}

	// node annotated with machine that does not exist
	nodeAnnotatedWithNoExistentMachine := mrotesting.NewNode("annotatedWithNoExistentMachine", true, "annotatedWithNoExistentMachine")

	// node annotated with machine without owner reference
	nodeAnnotatedWithMachineWithoutOwnerReference := mrotesting.NewNode("annotatedWithMachineWithoutOwnerReference", true, "machineWithoutOwnerController")
	machineWithoutOwnerController := mrotesting.NewMachine("machineWithoutOwnerController", nodeAnnotatedWithMachineWithoutOwnerReference.Name, "")
	machineWithoutOwnerController.OwnerReferences = nil

	// node annotated with machine without node reference
	nodeAnnotatedWithMachineWithoutNodeReference := mrotesting.NewNode("annotatedWithMachineWithoutNodeReference", true, "machineWithoutNodeRef")
	machineWithoutNodeRef := mrotesting.NewMachine("machineWithoutNodeRef", nodeAnnotatedWithMachineWithoutNodeReference.Name, "")
	machineWithoutNodeRef.Status.NodeRef = nil

	machineHealthCheck := mrotesting.NewMachineHealthCheck("machineHealthCheck")

	// remediationReboot
	nodeUnhealthyForTooLong := mrotesting.NewNode("nodeUnhealthyForTooLong", false, "machineUnhealthyForTooLong")
	machineUnhealthyForTooLong := mrotesting.NewMachine("machineUnhealthyForTooLong", nodeUnhealthyForTooLong.Name, "")

	// remediation disabled annotation

	nodeWithRemediationDisabled := mrotesting.NewNode("nodeWithRemediationDisabled", true, "machineWithRemediationDisabled")
	machineWithRemediationDisabled := mrotesting.NewMachine("machineWithRemediationDisabled", "node", "")
	machineWithRemediationDisabled.Annotations[disableRemediationAnotationKey] = "true"

	testsCases := []struct {
		machine             *mapiv1.Machine
		node                *corev1.Node
		remediationStrategy mrv1.RemediationStrategyType
		expected            expectedReconcile
	}{
		{
			machine: machineUnhealthyForTooLong,
			node:    nodeUnhealthyForTooLong,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  false,
			},
			remediationStrategy: mrv1.RemediationStrategyTypeReboot,
		},
		{
			machine: machineWithNodeHealthy,
			node:    nodeHealthy,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  false,
			},
		},
		{
			machine: machineWithNodeRecentlyUnhealthy,
			node:    nodeRecentlyUnhealthy,
			expected: expectedReconcile{
				result: reconcile.Result{
					Requeue:      true,
					RequeueAfter: remediationWaitTime,
				},
				error: false,
			},
		},
		{
			machine: nil,
			node:    nodeWithoutMachineAnnotation,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  false,
			},
		},
		{
			machine: nil,
			node:    nodeAnnotatedWithNoExistentMachine,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  false,
			},
		},
		{
			machine: machineWithoutOwnerController,
			node:    nodeAnnotatedWithMachineWithoutOwnerReference,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  false,
			},
		},
		{
			machine: machineWithoutNodeRef,
			node:    nodeAnnotatedWithMachineWithoutNodeReference,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  true,
			},
		},
		{
			machine: machineWithRemediationDisabled,
			node:    nodeWithRemediationDisabled,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  false,
			},
		},
	}

	for _, tc := range testsCases {
		machineHealthCheck.Spec.RemediationStrategy = &tc.remediationStrategy
		objects := []runtime.Object{}
		objects = append(objects, initObjects...)
		objects = append(objects, machineHealthCheck)
		if tc.machine != nil {
			objects = append(objects, tc.machine)
		}
		objects = append(objects, tc.node)
		r := newFakeReconciler(objects...)

		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: metav1.NamespaceNone,
				Name:      tc.node.Name,
			},
		}
		result, err := r.Reconcile(request)
		if tc.expected.error != (err != nil) {
			var errorExpectation string
			if !tc.expected.error {
				errorExpectation = "no"
			}
			t.Errorf("Test case: %s. Expected: %s error, got: %v", tc.node.Name, errorExpectation, err)
		}

		if result != tc.expected.result {
			if tc.expected.result.Requeue {
				before := tc.expected.result.RequeueAfter - time.Second
				after := tc.expected.result.RequeueAfter + time.Second
				if after < result.RequeueAfter || before > result.RequeueAfter {
					t.Errorf("Test case: %s. Expected RequeueAfter between: %v and %v, got: %v", tc.node.Name, before, after, result)
				}
			} else {
				t.Errorf("Test case: %s. Expected: %v, got: %v", tc.node.Name, tc.expected.result, result)
			}
		}
		if tc.remediationStrategy == mrv1.RemediationStrategyTypeReboot {
			machineRemediations := &mrv1.MachineRemediationList{}
			if err := r.client.List(context.TODO(), machineRemediations); err != nil {
				t.Errorf("Expected: no error, got: %v", err)
			}

			var mrExist bool
			for _, mr := range machineRemediations.Items {
				if mr.Spec.MachineName == tc.machine.Name {
					mrExist = true
				}
			}
			if !mrExist {
				t.Errorf("Expected: machine remediation with machine name %s should exist, got: %v", tc.machine.Name, machineRemediations.Items)
			}
		}
	}
}

func TestReconcileWithoutUnhealthyConditionsConfigMap(t *testing.T) {
	testReconcile(t, 5*time.Minute)
}

func TestReconcileWithUnhealthyConditionsConfigMap(t *testing.T) {
	cmBadConditions := mrotesting.NewUnhealthyConditionsConfigMap(mrv1.ConfigMapNodeUnhealthyConditions, badConditionsData)
	testReconcile(t, 1*time.Minute, cmBadConditions)
}

func TestHasMachineSetOwner(t *testing.T) {
	machineWithMachineSet := mrotesting.NewMachine("machineWithMachineSet", "node", "")
	machineWithNoMachineSet := mrotesting.NewMachine("machineWithNoMachineSet", "node", "")
	machineWithNoMachineSet.OwnerReferences = nil

	testsCases := []struct {
		machine  *mapiv1.Machine
		expected bool
	}{
		{
			machine:  machineWithNoMachineSet,
			expected: false,
		},
		{
			machine:  machineWithMachineSet,
			expected: true,
		},
	}

	for _, tc := range testsCases {
		if got := hasMachineSetOwner(tc.machine); got != tc.expected {
			t.Errorf("Test case: Machine %s. Expected: %t, got: %t", tc.machine.Name, tc.expected, got)
		}
	}

}

func TestUnhealthyForTooLong(t *testing.T) {
	nodeUnhealthyForTooLong := mrotesting.NewNode("nodeUnhealthyForTooLong", false, "")
	nodeRecentlyUnhealthy := mrotesting.NewNode("nodeRecentlyUnhealthy", false, "")
	nodeRecentlyUnhealthy.Status.Conditions[0].LastTransitionTime = metav1.Time{Time: time.Now()}
	testsCases := []struct {
		node     *corev1.Node
		expected bool
	}{
		{
			node:     nodeUnhealthyForTooLong,
			expected: true,
		},
		{
			node:     nodeRecentlyUnhealthy,
			expected: false,
		},
	}
	for _, tc := range testsCases {
		if got := unhealthyForTooLong(&tc.node.Status.Conditions[0], time.Minute); got != tc.expected {
			t.Errorf("Test case: %s. Expected: %t, got: %t", tc.node.Name, tc.expected, got)
		}
	}
}

func TestApplyRemediationReboot(t *testing.T) {
	nodeUnhealthyForTooLong := mrotesting.NewNode("nodeUnhealthyForTooLong", false, "machineUnhealthyForTooLong")
	machineUnhealthyForTooLong := mrotesting.NewMachine("machineUnhealthyForTooLong", nodeUnhealthyForTooLong.Name, "")
	machineHealthCheck := mrotesting.NewMachineHealthCheck("machineHealthCheck")
	r := newFakeReconciler(nodeUnhealthyForTooLong, machineUnhealthyForTooLong, machineHealthCheck)
	_, err := r.remediationStrategyReboot(machineUnhealthyForTooLong, nodeUnhealthyForTooLong)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	machineRemediations := &mrv1.MachineRemediationList{}
	if err := r.client.List(context.TODO(), machineRemediations); err != nil {
		t.Errorf("Expected: no error, got: %v", err)
	}

	var mrExist bool
	for _, mr := range machineRemediations.Items {
		if mr.Spec.MachineName == machineUnhealthyForTooLong.Name {
			mrExist = true
		}
	}
	if !mrExist {
		t.Errorf("Expected: machine remediation with machine name %s should exist, got: no machine remediations", machineUnhealthyForTooLong.Name)
	}
}

func TestGetNodeNamesForMHC(t *testing.T) {
	testCases := []struct {
		mhc               *mrv1.MachineHealthCheck
		machines          []*mapiv1.Machine
		expectedNodeNames []types.NodeName
	}{
		{
			mhc: mrotesting.NewMachineHealthCheck("matchNodes"),
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("test", "node1", ""),
				mrotesting.NewMachine("test2", "node2", ""),
			},
			expectedNodeNames: []types.NodeName{
				"node1",
				"node2",
			},
		},
		{
			mhc: &mrv1.MachineHealthCheck{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "noMatchingMachines",
					Namespace: consts.NamespaceOpenshiftMachineAPI,
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "MachineHealthCheck",
				},
				Spec: mrv1.MachineHealthCheckSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"no": "match",
						},
					},
				},
				Status: mrv1.MachineHealthCheckStatus{},
			},
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("test", "node1", ""),
				mrotesting.NewMachine("test2", "node2", ""),
			},
			expectedNodeNames: nil,
		},
		{
			mhc: mrotesting.NewMachineHealthCheck("noNodeRefs"),
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("test", "", ""),
				mrotesting.NewMachine("test2", "", ""),
			},
			expectedNodeNames: nil,
		},
	}
	for _, tc := range testCases {
		objects := []runtime.Object{}
		for i := range tc.machines {
			objects = append(objects, runtime.Object(tc.machines[i]))
		}
		r := newFakeReconciler(objects...)
		nodeNames, err := r.getNodeNamesForMHC(tc.mhc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(nodeNames, tc.expectedNodeNames) {
			t.Errorf("Expected: %v, got: %v", tc.expectedNodeNames, nodeNames)
		}
	}
}

func TestNodeRequestsFromMachineHealthCheck(t *testing.T) {
	testCases := []struct {
		mhc              *mrv1.MachineHealthCheck
		machines         []*mapiv1.Machine
		expectedRequests []reconcile.Request
	}{
		{
			mhc: mrotesting.NewMachineHealthCheck("matchNodes"),
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("test", "node1", ""),
				mrotesting.NewMachine("test2", "node2", ""),
			},
			expectedRequests: []reconcile.Request{
				{
					NamespacedName: client.ObjectKey{
						Name: string("node1"),
					},
				},
				{
					NamespacedName: client.ObjectKey{
						Name: string("node2"),
					},
				},
			},
		},
		{
			mhc: &mrv1.MachineHealthCheck{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "noMatchingMachines",
					Namespace: consts.NamespaceOpenshiftMachineAPI,
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "MachineHealthCheck",
				},
				Spec: mrv1.MachineHealthCheckSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"no": "match",
						},
					},
				},
				Status: mrv1.MachineHealthCheckStatus{},
			},
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("test", "node1", ""),
				mrotesting.NewMachine("test2", "node2", ""),
			},
			expectedRequests: nil,
		},
		{
			mhc: mrotesting.NewMachineHealthCheck("noNodeRefs"),
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("test", "", ""),
				mrotesting.NewMachine("test2", "", ""),
			},
			expectedRequests: nil,
		},
	}
	for _, tc := range testCases {
		objects := []runtime.Object{}
		for i := range tc.machines {
			objects = append(objects, runtime.Object(tc.machines[i]))
		}
		objects = append(objects, runtime.Object(tc.mhc))
		r := newFakeReconciler(objects...)
		o := handler.MapObject{
			Meta:   tc.mhc.GetObjectMeta(),
			Object: tc.mhc,
		}
		requests := r.nodeRequestsFromMachineHealthCheck(o)
		if !reflect.DeepEqual(requests, tc.expectedRequests) {
			t.Errorf("Expected: %v, got: %v", tc.expectedRequests, requests)
		}
	}
}

func testUpdateMHCStatus(t *testing.T, initObjects ...runtime.Object) {
	machineHealthCheck := mrotesting.NewMachineHealthCheck("machineHealthCheck")
	unhealthyConditions := []conditions.UnhealthyCondition{
		conditions.UnhealthyCondition{
			Name:   corev1.NodeReady,
			Status: corev1.ConditionUnknown,
		},
	}
	emptyUnhealthyConditions := []conditions.UnhealthyCondition{}
	nodeUnhealthyConditions := []corev1.NodeConditionType{
		corev1.NodeReady,
	}
	targetedConditions := []mrv1.TargetedCondition{
		mrv1.TargetedCondition{
			Name:   unhealthyConditions[0].Name,
			Status: unhealthyConditions[0].Status,
		},
	}
	emptyTargetedConditions := []mrv1.TargetedCondition(nil)
	key := client.ObjectKey{
		Namespace: machineHealthCheck.GetNamespace(),
		Name:      machineHealthCheck.GetName(),
	}
	testsCases := []struct {
		machines   []*mapiv1.Machine
		nodes      []*corev1.Node
		conditions []conditions.UnhealthyCondition
		expected   *mrv1.MachineHealthCheckStatus
	}{
		// no machine
		{
			machines:   []*mapiv1.Machine{},
			nodes:      []*corev1.Node{},
			conditions: emptyUnhealthyConditions,
			expected: &mrv1.MachineHealthCheckStatus{
				TotalHealthyMachines: 0,
				TargetedConditions:   emptyTargetedConditions,
				TargetedMachines:     []mrv1.TargetedMachine(nil),
			},
		},
		// single machine
		{
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("machine1", "node1", ""),
			},
			nodes: []*corev1.Node{
				mrotesting.NewNode("node1", true, "machine1"),
			},
			conditions: emptyUnhealthyConditions,
			expected: &mrv1.MachineHealthCheckStatus{
				TotalHealthyMachines: 1,
				TargetedConditions:   emptyTargetedConditions,
				TargetedMachines: []mrv1.TargetedMachine{
					mrotesting.NewTargetedMachine("machine1", true, nil),
				},
			},
		},
		// multiple machines
		{
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("machine1", "node1", ""),
				mrotesting.NewMachine("machine2", "node2", ""),
				mrotesting.NewMachine("machine3", "node3", ""),
			},
			nodes: []*corev1.Node{
				mrotesting.NewNode("node1", true, "machine1"),
				mrotesting.NewNode("node2", true, "machine2"),
				mrotesting.NewNode("node3", true, "machine3"),
			},
			conditions: emptyUnhealthyConditions,
			expected: &mrv1.MachineHealthCheckStatus{
				TotalHealthyMachines: 3,
				TargetedConditions:   emptyTargetedConditions,
				TargetedMachines: []mrv1.TargetedMachine{
					mrotesting.NewTargetedMachine("machine1", true, nil),
					mrotesting.NewTargetedMachine("machine2", true, nil),
					mrotesting.NewTargetedMachine("machine3", true, nil),
				},
			},
		},
		// some unhealthy machines
		{
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("machine1", "node1", ""),
				mrotesting.NewMachine("machine2", "node2", ""),
				mrotesting.NewMachine("machine3", "node3", ""),
				mrotesting.NewMachine("machine4", "node4", ""),
				mrotesting.NewMachine("machine5", "node5", ""),
			},
			nodes: []*corev1.Node{
				mrotesting.NewNode("node1", false, "machine1"),
				mrotesting.NewNode("node2", true, "machine2"),
				mrotesting.NewNode("node3", false, "machine3"),
				mrotesting.NewNode("node4", false, "machine4"),
				mrotesting.NewNode("node5", true, "machine5"),
			},
			conditions: unhealthyConditions,
			expected: &mrv1.MachineHealthCheckStatus{
				TotalHealthyMachines: 2,
				TargetedConditions:   targetedConditions,
				TargetedMachines: []mrv1.TargetedMachine{
					mrotesting.NewTargetedMachine("machine1", false, nodeUnhealthyConditions),
					mrotesting.NewTargetedMachine("machine2", true, nil),
					mrotesting.NewTargetedMachine("machine3", false, nodeUnhealthyConditions),
					mrotesting.NewTargetedMachine("machine4", false, nodeUnhealthyConditions),
					mrotesting.NewTargetedMachine("machine5", true, nil),
				},
			},
		},
		// some unhealthy machines without conditions
		{
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("machine1", "node1", ""),
				mrotesting.NewMachine("machine2", "node2", ""),
				mrotesting.NewMachine("machine3", "node3", ""),
				mrotesting.NewMachine("machine4", "node4", ""),
				mrotesting.NewMachine("machine5", "node5", ""),
			},
			nodes: []*corev1.Node{
				mrotesting.NewNode("node1", false, "machine1"),
				mrotesting.NewNode("node2", true, "machine2"),
				mrotesting.NewNode("node3", false, "machine3"),
				mrotesting.NewNode("node4", false, "machine4"),
				mrotesting.NewNode("node5", true, "machine5"),
			},
			conditions: emptyUnhealthyConditions,
			expected: &mrv1.MachineHealthCheckStatus{
				TotalHealthyMachines: 5,
				TargetedConditions:   emptyTargetedConditions,
				TargetedMachines: []mrv1.TargetedMachine{
					mrotesting.NewTargetedMachine("machine1", true, nil),
					mrotesting.NewTargetedMachine("machine2", true, nil),
					mrotesting.NewTargetedMachine("machine3", true, nil),
					mrotesting.NewTargetedMachine("machine4", true, nil),
					mrotesting.NewTargetedMachine("machine5", true, nil),
				},
			},
		},
		// machine with no node ref
		{
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("machine1", "node1", ""),
				mrotesting.NewMachine("machine2", "", ""),
			},
			nodes: []*corev1.Node{
				mrotesting.NewNode("node1", true, "machine1"),
			},
			conditions: emptyUnhealthyConditions,
			expected:   nil,
		},
		// machine with non existing node ref
		{
			machines: []*mapiv1.Machine{
				mrotesting.NewMachine("machine1", "node1", ""),
				mrotesting.NewMachine("machine2", "node2", ""),
			},
			nodes: []*corev1.Node{
				mrotesting.NewNode("node1", true, "machine1"),
			},
			conditions: emptyUnhealthyConditions,
			expected:   nil,
		},
	}
	for _, tc := range testsCases {
		objects := []runtime.Object{}
		objects = append(objects, initObjects...)
		objects = append(objects, machineHealthCheck)
		for _, machine := range tc.machines {
			objects = append(objects, machine)
		}
		for _, node := range tc.nodes {
			objects = append(objects, node)
		}
		fakeClient := fake.NewFakeClient(objects...)
		err := updateMHCStatus(fakeClient, machineHealthCheck, tc.conditions)
		if tc.expected == nil {
			if err == nil {
				t.Errorf("Test case: %v. Error expected but didn't ocure", tc)
			}
			continue
		}
		if err != nil && tc.expected != nil {
			t.Errorf("Test case: %v. Unexpected error: %v.", tc, err)
			continue
		}
		updatedMhc := &mrv1.MachineHealthCheck{}
		if err := fakeClient.Get(context.TODO(), key, updatedMhc); err != nil {
			t.Errorf("Test case: %v. Unable to get updated MHC: %v", tc, err)
			continue
		}
		if !reflect.DeepEqual(updatedMhc.Status, *tc.expected) {
			t.Errorf("Test case: %v. Expected status: %v. Real status: %v", tc, *tc.expected, updatedMhc.Status)
		}
	}
}

func TestUpdateMHCStatus(t *testing.T) {
	testUpdateMHCStatus(t)
}

func TestUpdateMHCStatusMachineWithNoMatchingLabels(t *testing.T) {
	// machine with no matching labels
	machineWithNoMatchingLabels := mrotesting.NewMachine("machineWithNoMatchingLabels", "nodeForMachineWithNoMatchingLabels", "")
	machineWithNoMatchingLabels.Labels = map[string]string{"foo": "bar2"}
	nodeForMachineWithNotMatchingLabels := mrotesting.NewNode("nodeForMachineWithNotMatchingLabels", true, machineWithNoMatchingLabels.Name)
	testUpdateMHCStatus(t, machineWithNoMatchingLabels, nodeForMachineWithNotMatchingLabels)
}
