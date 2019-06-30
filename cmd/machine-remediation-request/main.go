package main

import (
	"flag"
	"runtime"

	"k8s.io/klog"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/golang/glog"
	"github.com/metal3-io/machine-remediation-request-operator/pkg/apis/machineremediationrequest/v1alpha1"
	"github.com/metal3-io/machine-remediation-request-operator/pkg/controller"
	"github.com/metal3-io/machine-remediation-request-operator/pkg/controller/machineremediationrequest"
)

func printVersion() {
	glog.Infof("Go Version: %s", runtime.Version())
	glog.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func main() {
	watchNamespace := flag.String("namespace", "", "Namespace that the controller watches to reconcile machine-api objects. If unspecified, the controller watches for machine-api objects across all namespaces.")
	flag.Parse()
	printVersion()

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		glog.Fatal(err)
	}

	opts := manager.Options{}
	if *watchNamespace != "" {
		opts.Namespace = *watchNamespace
		klog.Infof("Watching machine-api objects only in namespace %q for reconciliation.", opts.Namespace)
	}
	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, opts)
	if err != nil {
		glog.Fatal(err)
	}

	glog.Infof("Registering Components.")

	// Setup Scheme for all resources
	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}
	if err := mapiv1.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}

	addControllers := []func(manager.Manager) error{machineremediationrequest.Add}

	// Setup all Controllers
	if err := controller.AddToManager(mgr, addControllers); err != nil {
		glog.Fatal(err)
	}

	glog.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		glog.Fatal(err)
	}
}
