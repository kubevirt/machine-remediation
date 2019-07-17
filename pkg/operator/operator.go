package operator

import (
	"context"

	"github.com/golang/glog"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	mrv1 "kubevirt.io/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	"kubevirt.io/machine-remediation-operator/pkg/operator/components"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var _ reconcile.Reconciler = &ReconcileMachineRemediationOperator{}

// ReconcileMachineRemediationOperator reconciles a MachineRemediationOperator object
type ReconcileMachineRemediationOperator struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	namespace string
}

// AddWithRemediator creates a new MachineRemediationOperator Controller and adds it to the Manager.
// The Manager will set fields on the Controller and start it when the Manager is started.
func AddWithRemediator(mgr manager.Manager, opts manager.Options) error {
	r, err := newReconciler(mgr, opts)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

func newReconciler(mgr manager.Manager, opts manager.Options) (reconcile.Reconciler, error) {
	return &ReconcileMachineRemediationOperator{
		client:    mgr.GetClient(),
		namespace: opts.Namespace,
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("machine-remediation-operator-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &mrv1.MachineRemediationOperator{}}, &handler.EnqueueRequestForObject{})
}

// Reconcile monitors MachineRemediationOperator and bring all machine remediation components to desired state
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMachineRemediationOperator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	glog.V(4).Infof("Reconciling MachineRemediationOperator triggered by %s/%s\n", request.Namespace, request.Name)

	// Get MachineRemediation from request
	mro := &mrv1.MachineRemediationOperator{}
	err := r.client.Get(context.TODO(), request.NamespacedName, mro)
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
	if mro.DeletionTimestamp != nil {
		if err := r.deleteComponents(); err != nil {

		}
	}

	if err := r.createOrUpdateComponents(mro); err != nil {

	}

	return reconcile.Result{}, nil
}

func (r *ReconcileMachineRemediationOperator) createOrUpdateComponents(mro *mrv1.MachineRemediationOperator) error {
	for _, component := range components.Components {
		if err := r.createOrUpdateServiceAccount(component, r.namespace); err != nil {
			return err
		}

		if err := r.createOrUpdateClusterRole(component); err != nil {
			return err
		}

		if err := r.createOrUpdateClusterRoleBinding(component, r.namespace); err != nil {
			return err
		}

		if err := r.createOrUpdateDeployment(component, r.namespace, mro.Spec.ImageRegistry, mro.Spec.ImageTag); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileMachineRemediationOperator) deleteComponents() error {
	for _, component := range components.Components {
		if err := r.deleteDeployment(component, r.namespace); err != nil {
			return err
		}

		if err := r.deleteClusterRoleBinding(component); err != nil {
			return err
		}

		if err := r.deleteClusterRole(component); err != nil {
			return err
		}

		if err := r.deleteServiceAccount(component, r.namespace); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileMachineRemediationOperator) getDeployment(name string, namespace string) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{}
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	if err := r.client.Get(context.TODO(), key, deploy); err != nil {
		return nil, err
	}
	return deploy, nil
}

func (r *ReconcileMachineRemediationOperator) createOrUpdateDeployment(name string, namespace string, imageRepository string, imageTag string) error {
	newDeploy := components.NewDeployment(name, namespace, imageRepository, imageTag)

	_, err := r.getDeployment(name, namespace)
	if errors.IsNotFound(err) {
		if err := r.client.Create(context.TODO(), newDeploy); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return r.client.Update(context.TODO(), newDeploy)
}

func (r *ReconcileMachineRemediationOperator) deleteDeployment(name string, namespace string) error {
	deploy, err := r.getDeployment(name, namespace)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return r.client.Delete(context.TODO(), deploy)
}

func (r *ReconcileMachineRemediationOperator) getServiceAccount(name string, namespace string) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	if err := r.client.Get(context.TODO(), key, sa); err != nil {
		return nil, err
	}
	return sa, nil
}

func (r *ReconcileMachineRemediationOperator) createOrUpdateServiceAccount(name string, namespace string) error {
	newServiceAccount := components.NewServiceAccount(name, namespace)

	_, err := r.getServiceAccount(name, namespace)
	if errors.IsNotFound(err) {
		if err := r.client.Create(context.TODO(), newServiceAccount); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return r.client.Update(context.TODO(), newServiceAccount)
}

func (r *ReconcileMachineRemediationOperator) deleteServiceAccount(name string, namespace string) error {
	sa, err := r.getServiceAccount(name, namespace)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return r.client.Delete(context.TODO(), sa)
}

func (r *ReconcileMachineRemediationOperator) getClusterRole(name string) (*rbacv1.ClusterRole, error) {
	cr := &rbacv1.ClusterRole{}
	key := types.NamespacedName{
		Name:      name,
		Namespace: metav1.NamespaceNone,
	}
	if err := r.client.Get(context.TODO(), key, cr); err != nil {
		return nil, err
	}
	return cr, nil
}

func (r *ReconcileMachineRemediationOperator) createOrUpdateClusterRole(name string) error {
	newClusterRole := components.NewClusterRole(name, components.Rules[name])

	_, err := r.getClusterRole(name)
	if errors.IsNotFound(err) {
		if err := r.client.Create(context.TODO(), newClusterRole); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return r.client.Update(context.TODO(), newClusterRole)
}

func (r *ReconcileMachineRemediationOperator) deleteClusterRole(name string) error {
	cr, err := r.getClusterRole(name)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return r.client.Delete(context.TODO(), cr)
}

func (r *ReconcileMachineRemediationOperator) getClusterRoleBinding(name string) (*rbacv1.ClusterRoleBinding, error) {
	crb := &rbacv1.ClusterRoleBinding{}
	key := types.NamespacedName{
		Name:      name,
		Namespace: metav1.NamespaceNone,
	}
	if err := r.client.Get(context.TODO(), key, crb); err != nil {
		return nil, err
	}
	return crb, nil
}

func (r *ReconcileMachineRemediationOperator) createOrUpdateClusterRoleBinding(name string, namespace string) error {
	newClusterRoleBinding := components.NewClusterRoleBinding(name, namespace)

	_, err := r.getClusterRoleBinding(name)
	if errors.IsNotFound(err) {
		if err := r.client.Create(context.TODO(), newClusterRoleBinding); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return r.client.Update(context.TODO(), newClusterRoleBinding)
}

func (r *ReconcileMachineRemediationOperator) deleteClusterRoleBinding(name string) error {
	crb, err := r.getClusterRoleBinding(name)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return r.client.Delete(context.TODO(), crb)
}
