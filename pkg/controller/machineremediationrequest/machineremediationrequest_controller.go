package machineremediationrequest

import (
	mrrv1 "github.com/openshift/machine-remediation-request-operator/pkg/apis/machineremediationrequest/v1alpha1"
	mrrutils "github.com/openshift/machine-remediation-request-operator/pkg/utils"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new MachineRemediationRequest Controller and adds it to the Manager. The Manager will set fields on the Controller
// and start it when the Manager is started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	r := &ReconcileMachineRemediationRequest{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}

	ns, err := mrrutils.GetNamespace(mrrutils.ServiceAccountNamespaceFile)
	if err != nil {
		return r, err
	}

	r.namespace = ns
	return r, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("machineremediationrequest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &mrrv1.MachineRemediationRequest{}}, &handler.EnqueueRequestForObject{})
}

var _ reconcile.Reconciler = &ReconcileMachineRemediationRequest{}

// ReconcileMachineRemediationRequest reconciles a MachineRemediationRequest object
type ReconcileMachineRemediationRequest struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	namespace string
}

// Reconcile monitors MachineRemediationRequest and apply the remediation strategy in the case when the
// MachineRemediationRequest was created
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMachineRemediationRequest) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
