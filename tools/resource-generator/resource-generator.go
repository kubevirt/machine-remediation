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

	"kubevirt.io/machine-remediation-operator/pkg/operator/components"
	"kubevirt.io/machine-remediation-operator/tools/utils"
)

func main() {
	// General arguments
	resourceType := flag.String("type", "", "Type of resource to generate.")
	namespace := flag.String("namespace", "kube-system", "Namespace to use.")
	version := flag.String("version", "latest", "version to use.")
	pullPolicy := flag.String("pullPolicy", "IfNotPresent", "ImagePullPolicy to use.")
	verbosity := flag.String("verbosity", "2", "Verbosity level to use.")

	// controllers images
	mdbImage := flag.String("mdb-image", "", "Machine disruption budget controller image, should include a repository and a tag.")
	mhcImage := flag.String("mhc-image", "", "Machine health check controller image, should include a repository and a tag.")
	mrImage := flag.String("mr-image", "", "Machine remediation controller image, should include a repository and a tag.")
	mroImage := flag.String("mro-image", "", "Machine remediation operator controller image, should include a repository and a tag.")

	flag.Parse()

	imagePullPolicy := corev1.PullPolicy(*pullPolicy)

	switch *resourceType {
	case "machine-remediation-operator":
		// create service account for the machine-remediation-operator
		sa := components.NewServiceAccount(*resourceType, *namespace, *version)
		utils.MarshallObject(sa, os.Stdout)

		// create cluster role for the machine-remediation-operator
		cr := components.NewClusterRole(*resourceType, components.Rules[*resourceType], *version)
		utils.MarshallObject(cr, os.Stdout)

		// create cluster role binding for the machine-remediation-operator
		crb := components.NewClusterRoleBinding(*resourceType, *namespace, *version)
		utils.MarshallObject(crb, os.Stdout)

		// create operator deployment
		operatorData := &components.DeploymentData{
			ImageName:       *mroImage,
			Name:            *resourceType,
			Namespace:       *namespace,
			PullPolicy:      imagePullPolicy,
			Verbosity:       *verbosity,
			OperatorVersion: *version,
		}
		controllersImages := &components.ControllersImages{
			MachineDisruptionBudget: *mdbImage,
			MachineHealthCheck:      *mhcImage,
			MachineRemediation:      *mrImage,
		}
		operator := components.NewOperatorDeployment(operatorData, controllersImages)
		utils.MarshallObject(operator, os.Stdout)
	case "machine-remediation-operator-cr":
		// create operator CR
		mro := components.NewMachineRemediationOperator(*resourceType, *namespace, imagePullPolicy, *version)
		mro.Name = "mro"
		utils.MarshallObject(mro, os.Stdout)
	default:
		panic(fmt.Errorf("unknown resource type %s", *resourceType))
	}
}
