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

package tests_test

import (
	"flag"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kubevirt.io/node-recovery/pkg/apis/noderecovery/v1alpha1"
	"kubevirt.io/node-recovery/pkg/client"
	"kubevirt.io/node-recovery/pkg/controller"
	"kubevirt.io/node-recovery/tests"
)

var _ = Describe("Node Remediation", func() {
	var sshExecPod *corev1.Pod

	flag.Parse()

	kubeClient := client.NewKubeClientSet()
	nrClient := client.NewNodeRecoveryClientSet()
	nodeConditionManager := controller.NewNodeConditionManager()

	BeforeEach(func() {
		By("Getting master node")
		masterNode, err := tests.GetMasterNode()
		Expect(err).ToNot(HaveOccurred())

		By("Creating fake IPMI server pod")
		_, err = tests.CreateFakeIpmiPod(masterNode.Name)
		Expect(err).ToNot(HaveOccurred())

		By("Creating service for fake IPMI server")
		_, err = tests.CreateFakeIpmiService(tests.ServiceFakeIpmiClusterIP, tests.ServiceFakeIpmiPort)
		Expect(err).ToNot(HaveOccurred())

		By("Creating pod to execute SSH commands")
		sshExecPod, err = tests.CreateSSHExecPod(masterNode.Name)
		Expect(err).ToNot(HaveOccurred())
	})

	When("node has \"NotReady\" state", func() {
		BeforeEach(func() {
			By("Stoping kubelet service on the non-master node")
			Eventually(
				func() error {
					_, _, err := tests.RunSSHCommand(
						sshExecPod,
						tests.NonMasterNode,
						tests.NodeUser,
						[]string{"sudo", "systemctl", "stop", "kubelet"},
					)
					return err
				}, 60*time.Second, time.Second,
			).ShouldNot(HaveOccurred())

			By("Waiting until non-master node will have \"NonReady\" state")
			Eventually(
				func() bool {
					node, err := kubeClient.CoreV1().Nodes().Get(tests.NonMasterNode, metav1.GetOptions{})
					Expect(err).ToNot(HaveOccurred())

					readyCond := nodeConditionManager.GetNodeCondition(node, corev1.NodeReady)
					return readyCond.Status == corev1.ConditionUnknown
				}, 90*time.Second, 5*time.Second,
			).Should(BeTrue())
		})

		It("should remediate the node", func() {
			By("Checking that node remediation object was created")
			Eventually(
				func() *v1.NodeRemediation {
					nr, err := nrClient.NoderecoveryV1alpha1().NodeRemediations().Get(tests.NonMasterNode, metav1.GetOptions{})
					if err != nil {
						return nil
					}
					return nr
				}, 30*time.Second, 1*time.Second,
			).ShouldNot(BeNil())
		})

		AfterEach(func() {
			By("Starting kubelet service on the non-master node")
			Eventually(
				func() error {
					_, _, err := tests.RunSSHCommand(
						sshExecPod,
						tests.NonMasterNode,
						tests.NodeUser,
						[]string{"sudo", "systemctl", "start", "kubelet"},
					)
					return err
				}, 60*time.Second, time.Second,
			).ShouldNot(HaveOccurred())

			By("Waiting until non-master node will have \"Ready\" state")
			Eventually(
				func() bool {
					node, err := kubeClient.CoreV1().Nodes().Get(tests.NonMasterNode, metav1.GetOptions{})
					Expect(err).ToNot(HaveOccurred())

					readyCond := nodeConditionManager.GetNodeCondition(node, corev1.NodeReady)
					return readyCond.Status == corev1.ConditionTrue
				}, 90*time.Second, 5*time.Second,
			).Should(BeTrue())
		})
	})
})
