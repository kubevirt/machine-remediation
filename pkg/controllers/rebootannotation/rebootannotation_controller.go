package rebootannotation

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	mrv1 "github.com/openshift/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	"github.com/openshift/machine-remediation-operator/pkg/consts"
	mrutils "github.com/openshift/machine-remediation-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ reconcile.Reconciler = &ReconcileRebootAnnotation{}

// ReconcileRebootAnnotation reconciles a node with reboot annotation
type ReconcileRebootAnnotation struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	namespace string
}

// Add creates a new RebootAnnotation Controller and adds it to the Manager.
// The Manager will set fields on the Controller and start it when the Manager is started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	namespace, err := mrutils.GetNamespace(mrutils.ServiceAccountNamespaceFile)
	if err != nil {
		return nil, err
	}

	return &ReconcileRebootAnnotation{
		client:    mgr.GetClient(),
		namespace: namespace,
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rebootannotation-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	p := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			node, ok := e.ObjectNew.(*corev1.Node)
			if !ok {
				return false
			}

			if _, ok = node.Annotations[consts.AnnotationReboot]; !ok {
				return false
			}
			return true
		},
	}

	return c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}, p)
}

// Reconcile monitors nodes and creates MachineRemediation object for each node with reboot annotation.
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRebootAnnotation) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	glog.V(4).Infof("Reconciling node triggered by %s/%s\n", request.Namespace, request.Name)

	// Get MachineRemediation from request
	node := &corev1.Node{}
	err := r.client.Get(context.TODO(), request.NamespacedName, node)
	glog.V(4).Infof("Reconciling, getting node %v", node)
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

	// Get machine object from the node annotation
	machine, err := getMachineByNode(r.client, node)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Creating the machine remediation resource
	mr := &mrv1.MachineRemediation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      machine.Name,
			Namespace: machine.Namespace,
		},
		Spec: mrv1.MachineRemediationSpec{
			MachineName: machine.Name,
			Type:        mrv1.RemediationTypeReboot,
		},
		Status: mrv1.MachineRemediationStatus{
			State:     mrv1.RemediationStateStarted,
			Reason:    "Machine remediation started",
			StartTime: &metav1.Time{Time: time.Now()},
		},
	}
	if err = r.client.Create(context.TODO(), mr); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// getMachineByNode returns the machine object that mapped to the node
func getMachineByNode(c client.Client, node *corev1.Node) (*mapiv1.Machine, error) {
	machineKey, ok := node.Annotations[consts.AnnotationMachine]
	if !ok {
		glog.Warningf("No machine annotation for node %s", node.Name)
		return nil, fmt.Errorf("No machine annotation for node %s", node.Name)
	}

	glog.V(4).Infof("Node %s is annotated with machine %s", node.Name, machineKey)
	machine := &mapiv1.Machine{}
	machineNamespace, machineName, err := cache.SplitMetaNamespaceKey(machineKey)
	if err != nil {
		return nil, err
	}

	key := types.NamespacedName{
		Namespace: machineNamespace,
		Name:      machineName,
	}
	if err := c.Get(context.TODO(), key, machine); err != nil {
		return nil, err
	}
	return machine, nil
}
