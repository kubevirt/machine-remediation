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
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/spf13/pflag"

	"kubevirt.io/machine-remediation-operator/pkg/operator/components"
	"kubevirt.io/machine-remediation-operator/tools/utils"
)

type templateData struct {
	Namespace              string
	ContainerTag           string
	ContainerPrefix        string
	ImagePullPolicy        string
	Verbosity              string
	OperatorDeploymentSpec string
	OperatorRules          string
	CsvVersion             string
	GeneratedManifests     map[string]string
}

func main() {
	namespace := flag.String("namespace", "", "")
	containerPrefix := flag.String("container-prefix", "", "")
	containerTag := flag.String("container-tag", "", "")
	imagePullPolicy := flag.String("image-pull-policy", "IfNotPresent", "")
	verbosity := flag.String("verbosity", "2", "")
	genDir := flag.String("generated-manifests-dir", "", "")
	inputFile := flag.String("input-file", "", "")
	processFiles := flag.Bool("process-files", false, "")
	processVars := flag.Bool("process-vars", false, "")
	csvVersion := flag.String("csv-version", "", "")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.Parse()

	if !(*processFiles || *processVars) {
		panic("at least one of process-files or process-vars must be true")
	}

	data := templateData{
		GeneratedManifests: make(map[string]string),
	}

	if *processVars {
		data.Namespace = *namespace
		data.ContainerTag = *containerTag
		data.ContainerPrefix = *containerPrefix
		data.ImagePullPolicy = *imagePullPolicy
		data.Verbosity = fmt.Sprintf("%s", *verbosity)
		data.OperatorDeploymentSpec = fixResourceString(getOperatorDeploymentSpec(data), 12)
		data.OperatorRules = fixResourceString(getOperatorRules(), 10)
		data.CsvVersion = *csvVersion
	} else {
		data.Namespace = "{{.Namespace}}"
		data.ContainerTag = "{{.ContainerTag}}"
		data.ContainerPrefix = "{{.ContainerPrefix}}"
		data.ImagePullPolicy = "{{.ImagePullPolicy}}"
		data.Verbosity = "{{.Verbosity}}"
		data.OperatorDeploymentSpec = "{{.OperatorDeploymentSpec}}"
		data.OperatorRules = "{{.OperatorRules}}"
		data.CsvVersion = "{{.CsvVersion}}"
	}

	if *processFiles {
		manifests, err := ioutil.ReadDir(*genDir)
		if err != nil {
			panic(err)
		}

		for _, manifest := range manifests {
			if manifest.IsDir() {
				continue
			}
			b, err := ioutil.ReadFile(filepath.Join(*genDir, manifest.Name()))
			if err != nil {
				panic(err)
			}
			data.GeneratedManifests[manifest.Name()] = string(b)
		}
	}

	tmpl := template.Must(template.ParseFiles(*inputFile))
	err := tmpl.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}

func getOperatorDeploymentSpec(data templateData) string {
	imagePullPolicy := corev1.PullPolicy(data.ImagePullPolicy)
	deployment := components.NewDeployment(components.ComponentMachineRemediationOperator, data.Namespace, data.ContainerPrefix, data.ContainerTag, imagePullPolicy, data.Verbosity)
	return objToStr(deployment.Spec)
}

type clusterRoleRules struct {
	Rules []rbacv1.PolicyRule `json:"rules"`
}

func getOperatorRules() string {
	rule := clusterRoleRules{
		Rules: components.Rules[components.ComponentMachineRemediationOperator],
	}
	return objToStr(rule)
}

func objToStr(obj interface{}) string {
	writer := strings.Builder{}
	err := utils.MarshallObject(obj, &writer)
	if err != nil {
		panic(err)
	}
	return writer.String()
}

func fixResourceString(in string, indention int) string {
	out := strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(in))
	for scanner.Scan() {
		line := scanner.Text()
		// remove separator lines
		if !strings.HasPrefix(line, "---") {
			// indent so that it fits into the manifest
			// spaces is is indention - 2, because we want to have 2 spaces less for being able to start an array
			spaces := strings.Repeat(" ", indention-2)
			if strings.HasPrefix(line, "apiGroups") {
				// spaces + array start
				out.WriteString(spaces + "- " + line + "\n")
			} else {
				// 2 more spaces
				out.WriteString(spaces + "  " + line + "\n")
			}
		}
	}
	return out.String()
}
