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
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"

	"kubevirt.io/node-recovery/tests"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

var _ = BeforeSuite(func() {
	ExpectWithOffset(1, tests.BeforeTestSuitSetup()).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	ExpectWithOffset(1, tests.AfterTestSuitCleanup()).ToNot(HaveOccurred())
	Eventually(
		func() bool {
			_, err := tests.GetNamespace(tests.NamespaceTest)
			return errors.IsNotFound(err)
		}, 180*time.Second, time.Second).Should(BeTrue())
})
