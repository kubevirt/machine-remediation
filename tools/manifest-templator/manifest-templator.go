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
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/pflag"
)

type templateData struct {
	ImageMachineRemediation string
	ImagePullPolicy         string
	Namespace               string
	Verbosity               string
	GeneratedManifests      map[string]string
}

func main() {
	namespace := flag.String("namespace", "", "")
	imagePullPolicy := flag.String("image-pull-policy", "IfNotPresent", "")
	verbosity := flag.String("verbosity", "2", "")
	genDir := flag.String("generated-manifests-dir", "", "")
	inputFile := flag.String("input-file", "", "")
	processFiles := flag.Bool("process-files", false, "")
	processVars := flag.Bool("process-vars", false, "")

	// controllers images
	mrImage := flag.String("mr-image", "", "Machine remediation controller image, should include a repository and a tag.")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.Parse()

	if !(*processFiles || *processVars) {
		panic("at least one of process-files or process-vars must be true")
	}

	data := &templateData{
		GeneratedManifests: make(map[string]string),
	}

	if *processVars {
		data.Namespace = *namespace
		data.ImageMachineRemediation = *mrImage
		data.ImagePullPolicy = *imagePullPolicy
		data.Verbosity = fmt.Sprintf("%s", *verbosity)
	} else {
		data.Namespace = "{{.Namespace}}"
		data.ImageMachineRemediation = "{{.ImageMachineRemediation}}"
		data.ImagePullPolicy = "{{.ImagePullPolicy}}"
		data.Verbosity = "{{.Verbosity}}"
	}

	if *processFiles {
		getGeneratedFiles(*genDir, data)
	}

	tmpl := template.Must(template.ParseFiles(*inputFile))
	err := tmpl.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}

func getGeneratedFiles(rootDir string, data *templateData) {
	manifests, err := ioutil.ReadDir(rootDir)
	if err != nil {
		panic(err)
	}

	for _, manifest := range manifests {
		if manifest.IsDir() {
			getGeneratedFiles(filepath.Join(rootDir, manifest.Name()), data)
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(rootDir, manifest.Name()))
		if err != nil {
			panic(err)
		}
		data.GeneratedManifests[manifest.Name()] = string(b)
	}
}
