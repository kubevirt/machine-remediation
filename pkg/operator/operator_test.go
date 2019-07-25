package operator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	mrv1 "kubevirt.io/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	mrotesting "kubevirt.io/machine-remediation-operator/pkg/utils/testing"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	imageRegistry = "docker.io/test"
	imageTag      = "test"
)

func init() {
	// Add types to scheme
	mrv1.AddToScheme(scheme.Scheme)
}

func verifyMachineRemediationOperatorConditions(
	conditions []mrv1.MachineRemediationOperatorStatusCondition,
	availabe mrv1.OperatorConditionStatus,
	degraded mrv1.OperatorConditionStatus,
	progressing mrv1.OperatorConditionStatus,
) bool {
	for _, c := range conditions {
		switch c.Type {
		case mrv1.OperatorAvailable:
			if c.Status != availabe {
				return false
			}
		case mrv1.OperatorDegraded:
			if c.Status != degraded {
				return false
			}
		case mrv1.OperatorProgressing:
			if c.Status != progressing {
				return false
			}
		}
	}
	return true
}

func newMachineRemediationOperator(name string) *mrv1.MachineRemediationOperator {
	return &mrv1.MachineRemediationOperator{
		TypeMeta: metav1.TypeMeta{Kind: "MachineRemediationOperator"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: mrotesting.NamespaceTest,
		},
		Spec: mrv1.MachineRemediationOperatorSpec{
			ImagePullPolicy: corev1.PullAlways,
			ImageRegistry:   imageRegistry,
			ImageTag:        imageTag,
		},
	}
}

// newFakeReconciler returns a new reconcile.Reconciler with a fake client
func newFakeReconciler(initObjects ...runtime.Object) *ReconcileMachineRemediationOperator {
	fakeClient := fake.NewFakeClient(initObjects...)
	return &ReconcileMachineRemediationOperator{
		client:    fakeClient,
		namespace: mrotesting.NamespaceTest,
	}
}

func TestReconcile(t *testing.T) {
	mro := newMachineRemediationOperator("mro")

	r := newFakeReconciler(mro)
	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: mrotesting.NamespaceTest,
			Name:      mro.Name,
		},
	}
	// first call to reconcile should only add the finalizer to the mro object
	result, err := r.Reconcile(request)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)

	updatedMro := &mrv1.MachineRemediationOperator{}
	key := types.NamespacedName{
		Name:      mro.Name,
		Namespace: mrotesting.NamespaceTest,
	}
	assert.NoError(t, r.client.Get(context.TODO(), key, updatedMro))
	assert.Equal(t, true, hasFinalizer(updatedMro))

	// second call to reconcile should create all componenets and update the status to progressing
	result, err = r.Reconcile(request)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, result)

	deploys := &appsv1.DeploymentList{}
	assert.NoError(t, r.client.List(context.TODO(), deploys))
	assert.Equal(t, 3, len(deploys.Items))
	for _, d := range deploys.Items {
		container := d.Spec.Template.Spec.Containers[0]
		assert.Equal(t, corev1.PullAlways, container.ImagePullPolicy)
		assert.Equal(t, fmt.Sprintf("%s/%s:%s", imageRegistry, container.Name, imageTag), container.Image)
	}

	updatedMro = &mrv1.MachineRemediationOperator{}
	assert.NoError(t, r.client.Get(context.TODO(), key, updatedMro))
	assert.Equal(t, true, verifyMachineRemediationOperatorConditions(
		updatedMro.Status.Conditions,
		mrv1.ConditionFalse,
		mrv1.ConditionFalse,
		mrv1.ConditionTrue,
	))

	// update all deployments status to have desired number of replicas
	for _, d := range deploys.Items {
		d.Status.Replicas = 1
		d.Status.UpdatedReplicas = 1
		assert.NoError(t, r.client.Update(context.TODO(), &d))
	}

	// third call to reconcile should set the operator status to available
	result, err = r.Reconcile(request)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)

	updatedMro = &mrv1.MachineRemediationOperator{}
	assert.NoError(t, r.client.Get(context.TODO(), key, updatedMro))
	assert.Equal(t, true, verifyMachineRemediationOperatorConditions(
		updatedMro.Status.Conditions,
		mrv1.ConditionTrue,
		mrv1.ConditionFalse,
		mrv1.ConditionFalse,
	))

	// update mro object deletion timestamp
	updatedMro.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	assert.NoError(t, r.client.Update(context.TODO(), updatedMro))

	result, err = r.Reconcile(request)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)

	deploys = &appsv1.DeploymentList{}
	assert.NoError(t, r.client.List(context.TODO(), deploys))
	assert.Equal(t, 0, len(deploys.Items))

	updatedMro = &mrv1.MachineRemediationOperator{}
	assert.NoError(t, r.client.Get(context.TODO(), key, updatedMro))
	assert.Equal(t, false, hasFinalizer(updatedMro))
}
