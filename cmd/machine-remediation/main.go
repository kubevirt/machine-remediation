package main

import (
	"flag"
	"runtime"

	"github.com/golang/glog"
	bmov1 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	mrv1 "github.com/openshift/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	"github.com/openshift/machine-remediation-operator/pkg/baremetal/remediator"
	"github.com/openshift/machine-remediation-operator/pkg/controller"
	"github.com/openshift/machine-remediation-operator/pkg/controller/machineremediation"
	"github.com/openshift/machine-remediation-operator/pkg/version"

	"k8s.io/klog"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func printVersion() {
	glog.Infof("Go Version: %s", runtime.Version())
	glog.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	glog.Infof("Component version: %s", version.Get())
}

func main() {
	watchNamespace := flag.String("namespace", "", "Namespace that the controller watches to reconcile machineremediation objects. If unspecified, the controller watches for machine-api objects across all namespaces.")
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
		klog.Infof("Watching machineremediation objects only in namespace %q for reconciliation.", opts.Namespace)
	}
	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, opts)
	if err != nil {
		glog.Fatal(err)
	}

	glog.Infof("Registering Components.")

	// Setup Scheme for all resources
	if err := mrv1.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}
	if err := mapiv1.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}
	if err := bmov1.SchemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		glog.Fatal(err)
	}

	bareMetalRemediator := remediator.NewBareMetalRemediator(mgr)
	addControllers := []func(manager.Manager) error{
		func(m manager.Manager) error { return machineremediation.AddWithRemediator(m, bareMetalRemediator) },
	}

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
