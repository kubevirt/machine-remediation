package utils

import (
	"context"
	"time"

	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"kubevirt.io/machine-remediation-operator/pkg/utils/conditions"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetWorkerNodes returns all nodes with the nodeWorkerRoleLabel label
func GetWorkerNodes(c client.Client) ([]corev1.Node, error) {
	workerNodes := &corev1.NodeList{}
	if err := c.List(context.TODO(), workerNodes, client.InNamespace(NamespaceOpenShiftMachineAPI), client.MatchingLabels(map[string]string{WorkerNodeRoleLabel: ""})); err != nil {
		return nil, err
	}
	return workerNodes.Items, nil
}

// FilterReadyNodes fileter the list of nodes and returns the list with ready nodes
func FilterReadyNodes(nodes []corev1.Node) []corev1.Node {
	var readyNodes []corev1.Node
	for _, n := range nodes {
		if IsNodeReady(&n) {
			readyNodes = append(readyNodes, n)
		}
	}
	return readyNodes
}

// IsNodeReady returns true once node is ready
func IsNodeReady(node *corev1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// WaitForNodeCondition waits for the node condition with the desired status
func WaitForNodeCondition(c client.Client, nodeName string, conditionType corev1.NodeConditionType, conditionStatus corev1.ConditionStatus, timeout time.Duration) error {
	key := types.NamespacedName{
		Name:      nodeName,
		Namespace: metav1.NamespaceNone,
	}
	glog.Infof("Wait until node %s will have %s condition with the status %s", nodeName, conditionType, conditionStatus)
	if err := wait.Poll(time.Second*10, timeout, func() (bool, error) {
		node := &corev1.Node{}
		if err := c.Get(context.TODO(), key, node); err != nil {
			return false, err
		}
		return conditions.NodeHasCondition(node, conditionType, conditionStatus), nil
	}); err != nil {
		return err
	}

	return nil
}

// WaitForNodesToGetReady waits until the cluster will have specified number of ready nodes
func WaitForNodesToGetReady(c client.Client, nodeLabels map[string]string, numberOfReadyNodes int, timeout time.Duration) error {
	glog.V(2).Infof("Wait until the environment will have %d ready nodes", numberOfReadyNodes)
	if err := wait.Poll(time.Second*10, timeout, func() (bool, error) {
		nodes := &corev1.NodeList{}
		if err := c.List(context.TODO(), nodes, client.MatchingLabels(nodeLabels)); err != nil {
			return false, err
		}

		readyNodes := 0
		for _, n := range nodes.Items {
			if conditions.NodeHasCondition(&n, corev1.NodeReady, corev1.ConditionTrue) {
				readyNodes++
			}
		}
		return readyNodes == numberOfReadyNodes, nil
	}); err != nil {
		return err
	}

	return nil
}
