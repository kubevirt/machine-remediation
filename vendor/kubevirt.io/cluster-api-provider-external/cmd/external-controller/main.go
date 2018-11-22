/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"sigs.k8s.io/cluster-api/pkg/apis"
	"sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	clustercontroller "sigs.k8s.io/cluster-api/pkg/controller/cluster"
	machinecontroller "sigs.k8s.io/cluster-api/pkg/controller/machine"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"kubevirt.io/cluster-api-provider-external/pkg/external"
)

func main() {
	var machineSetupConfigPath = "/etc/machinesetup/machine_setup_configs.yaml"
	flag.StringVar(&machineSetupConfigPath, "machinesetup", machineSetupConfigPath, "path to machine setup config file")

	flag.Set("logtostderr", "true")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	flag.Parse()

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		glog.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		glog.Fatal(err)
	}

	glog.Info("Registering Components")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}

	clusterClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}

	// Setup cluster controller
	clusterActuator, err := external.NewClusterActuator(clusterClient)
	if err != nil {
		panic(err)
	}
	clustercontroller.AddWithActuator(mgr, clusterActuator)

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	// Setup machine controller
	machineActuator, err := external.NewMachineActuator(kubeClient, clusterClient, machineSetupConfigPath)
	if err != nil {
		panic(err)
	}
	machinecontroller.AddWithActuator(mgr, machineActuator)

	// Setup node controller
	// TODO: enable when CR subresources available by default
	// nodecontroller.Add(mgr)

	glog.Info("Starting the manager")

	// Start the Cmd
	glog.Fatal(mgr.Start(signals.SetupSignalHandler()))
}
