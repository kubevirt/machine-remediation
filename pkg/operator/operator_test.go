package operator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	mrv1 "kubevirt.io/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	"kubevirt.io/machine-remediation-operator/pkg/consts"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	imageRegistry = "docker.io/test"
	imageTag      = "test"
)

func init() {
	// Add types to scheme
	extv1beta1.AddToScheme(scheme.Scheme)
	mrv1.AddToScheme(scheme.Scheme)
}

func verifyMachineRemediationOperatorConditions(
	conditions []mrv1.MachineRemediationOperatorStatusCondition,
	availabe corev1.ConditionStatus,
	degraded corev1.ConditionStatus,
	progressing corev1.ConditionStatus,
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
			Namespace: consts.NamespaceOpenshiftMachineAPI,
		},
		Spec: mrv1.MachineRemediationOperatorSpec{
			ImagePullPolicy: corev1.PullAlways,
			ImageRegistry:   imageRegistry,
		},
	}
}

// newFakeReconciler returns a new reconcile.Reconciler with a fake client
func newFakeReconciler(initObjects ...runtime.Object) *ReconcileMachineRemediationOperator {
	fakeClient := fake.NewFakeClient(initObjects...)
	return &ReconcileMachineRemediationOperator{
		client:           fakeClient,
		namespace:        consts.NamespaceOpenshiftMachineAPI,
		operatorVersion:  imageTag,
		crdsManifestsDir: "../../manifests/generated/crds",
	}
}

func TestReconcile(t *testing.T) {
	mro := newMachineRemediationOperator("mro")

	r := newFakeReconciler(mro)
	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: consts.NamespaceOpenshiftMachineAPI,
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
		Namespace: consts.NamespaceOpenshiftMachineAPI,
	}
	assert.NoError(t, r.client.Get(context.TODO(), key, updatedMro))
	assert.Equal(t, true, hasFinalizer(updatedMro))

	// second call to reconcile should create all componenets and update the status to progressing
	result, err = r.Reconcile(request)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, result)

	// verify that operator created all deployments
	deploys := &appsv1.DeploymentList{}
	assert.NoError(t, r.client.List(context.TODO(), deploys))
	assert.Equal(t, 3, len(deploys.Items))
	for _, d := range deploys.Items {
		container := d.Spec.Template.Spec.Containers[0]
		assert.Equal(t, corev1.PullAlways, container.ImagePullPolicy)
		assert.Equal(t, fmt.Sprintf("%s/%s:%s", imageRegistry, container.Name, imageTag), container.Image)
	}

	// verify that operator created all crds
	crds := &extv1beta1.CustomResourceDefinitionList{}
	assert.NoError(t, r.client.List(context.TODO(), crds))
	assert.Equal(t, 3, len(crds.Items))

	updatedMro = &mrv1.MachineRemediationOperator{}
	assert.NoError(t, r.client.Get(context.TODO(), key, updatedMro))
	assert.Equal(t, true, verifyMachineRemediationOperatorConditions(
		updatedMro.Status.Conditions,
		corev1.ConditionFalse,
		corev1.ConditionFalse,
		corev1.ConditionTrue,
	))

	// update all deployments status to have desired number of replicas
	for _, d := range deploys.Items {
		replicas, err := r.getReplicasCount()
		assert.NoError(t, err)

		d.Status.Replicas = replicas
		d.Status.UpdatedReplicas = replicas
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
		corev1.ConditionTrue,
		corev1.ConditionFalse,
		corev1.ConditionFalse,
	))

	// update mro object deletion timestamp
	updatedMro.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	assert.NoError(t, r.client.Update(context.TODO(), updatedMro))

	// verify that operator deletes all resources once it has deletion timestamp
	result, err = r.Reconcile(request)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)

	deploys = &appsv1.DeploymentList{}
	assert.NoError(t, r.client.List(context.TODO(), deploys))
	assert.Equal(t, 0, len(deploys.Items))

	crds = &extv1beta1.CustomResourceDefinitionList{}
	assert.NoError(t, r.client.List(context.TODO(), crds))
	assert.Equal(t, 0, len(crds.Items))

	updatedMro = &mrv1.MachineRemediationOperator{}
	assert.NoError(t, r.client.Get(context.TODO(), key, updatedMro))
	assert.Equal(t, false, hasFinalizer(updatedMro))
}
