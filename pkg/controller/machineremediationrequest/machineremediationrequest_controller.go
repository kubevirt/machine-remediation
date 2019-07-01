package machineremediationrequest

import (
	"context"

	"github.com/golang/glog"
	mrrv1 "github.com/openshift/machine-remediation-request-operator/pkg/apis/machineremediationrequest/v1alpha1"
	mrrutils "github.com/openshift/machine-remediation-request-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ reconcile.Reconciler = &ReconcileMachineRemediationRequest{}

// ReconcileMachineRemediationRequest reconciles a MachineRemediationRequest object
type ReconcileMachineRemediationRequest struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	remediator Remediator
	namespace  string
}

// AddWithRemediator creates a new MachineRemediationRequest Controller with remediator and adds it to the Manager.
// The Manager will set fields on the Controller and start it when the Manager is started.
func AddWithRemediator(mgr manager.Manager, remediator Remediator) error {
	r, err := newReconciler(mgr, remediator)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

func newReconciler(mgr manager.Manager, remediator Remediator) (reconcile.Reconciler, error) {
	namespace, err := mrrutils.GetNamespace(mrrutils.ServiceAccountNamespaceFile)
	if err != nil {
		return nil, err
	}

	return &ReconcileMachineRemediationRequest{
		client:     mgr.GetClient(),
		remediator: remediator,
		namespace:  namespace,
	}, nil
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

// Reconcile monitors MachineRemediationRequest and apply the remediation strategy in the case when the
// MachineRemediationRequest was created
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMachineRemediationRequest) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	glog.Infof("Reconciling MachineRemediationRequest triggered by %s/%s\n", request.Namespace, request.Name)

	// Get node from request
	mrr := &mrrv1.MachineRemediationRequest{}
	err := r.client.Get(context.TODO(), request.NamespacedName, mrr)
	glog.V(4).Infof("Reconciling, getting MachineRemediationRequest %v", mrr.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// we do not want to do anything on delete objects
	if mrr.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}
