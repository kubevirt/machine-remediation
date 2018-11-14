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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/node-recovery/pkg/client"
	"kubevirt.io/node-recovery/tests"
)

var _ = Describe("Node Remediation", func() {
	flag.Parse()
	kubeClient := client.NewKubeClientSet()

	It("test", func() {
		_, err := kubeClient.CoreV1().Pods("default").List(metav1.ListOptions{})
		Expect(err).ToNot(HaveOccurred())

		tests.ExecSSHCommand()
	})
})
