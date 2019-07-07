package remediator

import (
	"context"
	"fmt"
	"time"

	bmov1 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	mrv1 "github.com/openshift/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AnnotationBareMetalHost contains key for the bare metal host annotation
	AnnotationBareMetalHost = "metal3.io/BareMetalHost"
	// RebootDefaultTimeout contains minutes until the reboot will fail on the timeout
	RebootDefaultTimeout = 5
)

// BareMetalRemediator implements Remediator interface for bare metal machines
type BareMetalRemediator struct {
	client client.Client
}

// NewBareMetalRemediator returns new BareMetalRemediator object
func NewBareMetalRemediator() *BareMetalRemediator {
	return &BareMetalRemediator{}
}

// Recreate recreates the bare metal machine under the cluster
func (bmr *BareMetalRemediator) Recreate(ctx context.Context, machineRemediation *mrv1.MachineRemediation) error {
	return fmt.Errorf("Not implemented yet")
}

// Reboot reboots the bare metal machine
func (bmr *BareMetalRemediator) Reboot(ctx context.Context, machineRemediation *mrv1.MachineRemediation) error {
	// Get the machine from the MachineRemediation
	key := types.NamespacedName{
		Namespace: machineRemediation.Namespace,
		Name:      machineRemediation.Spec.MachineName,
	}
	machine := &mapiv1.Machine{}
	if err := bmr.client.Get(context.TODO(), key, machine); err != nil {
		return err
	}

	// Get the bare metal host object
	bmh, err := GetBareMetalHostByMachine(bmr.client, machine)
	if err != nil {
		return err
	}

	// Copy the BareMetalHost object to prevent modification of the original one
	newBmh := bmh.DeepCopy()

	// Copy the MachineRemediation object to prevent modification of the original one
	newMachineRemediation := machineRemediation.DeepCopy()

	var reason string
	var state mrv1.RemediationState

	switch *machineRemediation.Status.State {
	// initiating the reboot action
	case "":
		if !bmh.Spec.Online {
			state = mrv1.RemediationStateSucceeded
			reason = "Skipping the reboot, the machine power off by an user"
		} else {
			// power off the machine
			newBmh.Spec.Online = false
			if err := bmr.client.Update(context.TODO(), newBmh); err != nil {
				return err
			}
			state = mrv1.RemediationStateStarted
			reason = "Starts the reboot process"
		}

		// update the status of the MachineRemediation object
		newMachineRemediation.Status.StartTime = &metav1.Time{Time: time.Now()}
		newMachineRemediation.Status.State = &state
		newMachineRemediation.Status.Reason = &reason
		return bmr.client.Update(context.TODO(), newMachineRemediation)
	case mrv1.RemediationStateStarted:
		state = mrv1.RemediationStateInProgress
		reason = "Reboot in progress"

		// Reboot operation took more than defined timeout
		if machineRemediation.Status.StartTime.Time.Add(RebootDefaultTimeout * time.Minute).Before(time.Now()) {
			state = mrv1.RemediationStateFailed
			reason = "Reboot failed on timeout"
		}

		// host still has state on, we need to reconcile
		if bmh.Status.PoweredOn {
			return fmt.Errorf("machine %s still has power on state equal to true", machine.Name)
		}

		// power on the machine
		newBmh.Spec.Online = true
		if err := bmr.client.Update(context.TODO(), newBmh); err != nil {
			return err
		}

		// update the status of the MachineRemediation object
		newMachineRemediation.Status.State = &state
		newMachineRemediation.Status.Reason = &reason
		return bmr.client.Update(context.TODO(), newMachineRemediation)
	case mrv1.RemediationStateInProgress:
		// Reboot operation took more than defined timeout
		if machineRemediation.Status.StartTime.Time.Add(RebootDefaultTimeout * time.Minute).Before(time.Now()) {
			state = mrv1.RemediationStateFailed
			reason = "Reboot failed on timeout"
		}

		node, err := GetNodeByMachine(bmr.client, machine)
		if err != nil {
			return err
		}

		// Node back to Ready under the cluster
		if NodeHasCondition(node, corev1.NodeReady, corev1.ConditionTrue) {
			state = mrv1.RemediationStateSucceeded
			reason = "Reboot succeeded"
		}
		// update the status of the MachineRemediation object
		newMachineRemediation.Status.State = &state
		newMachineRemediation.Status.Reason = &reason
		return bmr.client.Update(context.TODO(), newMachineRemediation)
	}

	return nil
}

// GetBareMetalHostByMachine returns the bare metal host that linked to the machine
func GetBareMetalHostByMachine(c client.Client, machine *mapiv1.Machine) (*bmov1.BareMetalHost, error) {
	bmhKey, ok := machine.Annotations[AnnotationBareMetalHost]
	if !ok {
		return nil, fmt.Errorf("machine does not have bare metal host annotation")
	}

	bmhNamespace, bmhName, err := cache.SplitMetaNamespaceKey(bmhKey)
	bmh := &bmov1.BareMetalHost{}
	key := client.ObjectKey{
		Name:      bmhName,
		Namespace: bmhNamespace,
	}

	err = c.Get(context.TODO(), key, bmh)
	if err != nil {
		return nil, err
	}
	return bmh, nil
}

// NodeHasCondition returns true when the node has condition of the specific type and status
func NodeHasCondition(node *corev1.Node, conditionType corev1.NodeConditionType, contidionStatus corev1.ConditionStatus) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == conditionType && cond.Status == contidionStatus {
			return true
		}
	}
	return false
}

// GetNodeByMachine returns the node object referenced by machine
func GetNodeByMachine(c client.Client, machine *mapiv1.Machine) (*corev1.Node, error) {
	if machine.Status.NodeRef == nil {
		return nil, fmt.Errorf("machine %s does not have node reference", machine.Name)
	}

	node := &corev1.Node{}
	key := client.ObjectKey{
		Name:      machine.Status.NodeRef.Name,
		Namespace: machine.Status.NodeRef.Namespace,
	}

	err := c.Get(context.TODO(), key, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}
