package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	mrv1 "kubevirt.io/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	"kubevirt.io/machine-remediation-operator/pkg/utils/conditions"
	testsutils "kubevirt.io/machine-remediation-operator/tests/utils"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("[Feature:MachineRemediation]", func() {
	var c client.Client
	var numberOfReadyNodes int
	var testNode *corev1.Node

	stopKubeletAndValidateMachineReboot := func(testNode *corev1.Node, timeout time.Duration) {
		By(fmt.Sprintf("Stopping kubelet service on the node %s", testNode.Name))
		err := testsutils.StopKubelet(testNode.Name)
		Expect(err).ToNot(HaveOccurred())

		By(fmt.Sprintf("Validating that node %s has 'NotReady' condition", testNode.Name))
		err = testsutils.WaitForNodeCondition(c, testNode.Name, corev1.NodeReady, corev1.ConditionUnknown, testsutils.WaitLong)
		Expect(err).ToNot(HaveOccurred())

		By(fmt.Sprintf("Validating that node %s is deleted", testNode.Name))
		key := types.NamespacedName{
			Namespace: testNode.Namespace,
			Name:      testNode.Name,
		}
		Eventually(func() bool {
			node := &corev1.Node{}
			err := c.Get(context.TODO(), key, node)
			if err != nil {
				if errors.IsNotFound(err) {
					return true
				}
			}
			return false
		}, timeout, 5*time.Second).Should(BeTrue())
	}

	BeforeEach(func() {
		var err error
		c, err = testsutils.LoadClient()
		Expect(err).ToNot(HaveOccurred())

		nodes := &corev1.NodeList{}
		err = c.List(context.TODO(), nodes)
		Expect(err).ToNot(HaveOccurred())

		readyNodes := testsutils.FilterReadyNodes(nodes.Items)
		Expect(readyNodes).ToNot(BeEmpty())

		numberOfReadyNodes = len(readyNodes)
		testNode = &readyNodes[rand.Intn(numberOfReadyNodes)]
		glog.V(2).Infof("Test node %s", testNode.Name)

		testMachine, err := testsutils.GetMachineFromNode(c, testNode)
		Expect(err).ToNot(HaveOccurred())
		glog.V(2).Infof("Test machine %s", testMachine.Name)

		glog.V(2).Infof("Create machine health check with label selector: %s", testMachine.Labels)
		err = testsutils.CreateMachineHealthCheck(
			testsutils.MachineHealthCheckName,
			mrv1.RemediationStrategyTypeReboot,
			testMachine.Labels,
		)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("with node-unhealthy-conditions configmap", func() {
		BeforeEach(func() {
			unhealthyConditions := &conditions.UnhealthyConditions{
				Items: []conditions.UnhealthyCondition{
					{
						Name:    "Ready",
						Status:  "Unknown",
						Timeout: "60s",
					},
				},
			}
			glog.V(2).Infof("Create node-unhealthy-conditions configmap")
			err := testsutils.CreateUnhealthyConditionsConfigMap(unhealthyConditions)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should delete unhealthy machine", func() {
			stopKubeletAndValidateMachineReboot(testNode, 2*time.Minute)
		})

		AfterEach(func() {
			glog.V(2).Infof("Delete node-unhealthy-conditions configmap")
			err := testsutils.DeleteUnhealthyConditionsConfigMap()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	AfterEach(func() {
		testsutils.WaitForNodesToGetReady(c, map[string]string{}, numberOfReadyNodes, 15*time.Minute)
		testsutils.DeleteMachineHealthCheck(testsutils.MachineHealthCheckName)
		testsutils.DeleteKubeletKillerPods()
	})
})
