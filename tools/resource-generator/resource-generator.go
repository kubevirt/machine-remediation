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

package main

import (
	"flag"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"

	"kubevirt.io/machine-remediation/pkg/components"
	"kubevirt.io/machine-remediation/tools/utils"
)

func main() {
	// General arguments
	resourceType := flag.String("type", "", "Type of resource to generate.")
	namespace := flag.String("namespace", "kube-system", "Namespace to use.")
	pullPolicy := flag.String("pullPolicy", "IfNotPresent", "ImagePullPolicy to use.")
	verbosity := flag.String("verbosity", "2", "Verbosity level to use.")

	// controllers images
	mrImage := flag.String("mr-image", "", "Machine remediation controller image, should include a repository and a tag.")

	flag.Parse()

	imagePullPolicy := corev1.PullPolicy(*pullPolicy)

	switch *resourceType {
	case "machine-remediation":
		// create service account for the machine-remediation
		sa := components.NewServiceAccount(*resourceType, *namespace)
		utils.MarshallObject(sa, os.Stdout)

		// create cluster role for the machine-remediation
		cr := components.NewClusterRole(*resourceType, components.Rules[*resourceType])
		utils.MarshallObject(cr, os.Stdout)

		// create cluster role binding for the machine-remediation
		crb := components.NewClusterRoleBinding(*resourceType, *namespace)
		utils.MarshallObject(crb, os.Stdout)

		// create operator deployment
		deployData := &components.DeploymentData{
			ImageName:  *mrImage,
			Name:       *resourceType,
			Namespace:  *namespace,
			PullPolicy: imagePullPolicy,
			Verbosity:  *verbosity,
		}
		deploy := components.NewDeployment(deployData)
		utils.MarshallObject(deploy, os.Stdout)
	default:
		panic(fmt.Errorf("unknown resource type %s", *resourceType))
	}
}
