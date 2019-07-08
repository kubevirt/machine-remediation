package remediator

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	bmov1 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
	mrv1 "github.com/openshift/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	annotationBareMetalHost = "metal3.io/BareMetalHost"
	annotationReboot        = "metal3.io/BareMetalHostReboot"
	rebootDefaultTimeout    = 5
)

// BareMetalRemediator implements Remediator interface for bare metal machines
type BareMetalRemediator struct {
	client client.Client
}

// NewBareMetalRemediator returns new BareMetalRemediator object
func NewBareMetalRemediator(mgr manager.Manager) *BareMetalRemediator {
	return &BareMetalRemediator{}
}

// Recreate recreates the bare metal machine under the cluster
func (bmr *BareMetalRemediator) Recreate(ctx context.Context, machineRemediation *mrv1.MachineRemediation) error {
	return fmt.Errorf("Not implemented yet")
}

// Reboot reboots the bare metal machine
func (bmr *BareMetalRemediator) Reboot(ctx context.Context, machineRemediation *mrv1.MachineRemediation) error {
	glog.V(4).Infof("MachineRemediation %s has state %s", machineRemediation.Name, machineRemediation.Status.State)

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

	node, err := GetNodeByMachine(bmr.client, machine)
	if err != nil {
		return err
	}

	// Copy the BareMetalHost object to prevent modification of the original one
	newBmh := bmh.DeepCopy()

	// Copy the MachineRemediation object to prevent modification of the original one
	newMachineRemediation := machineRemediation.DeepCopy()

	var reason string
	var state mrv1.RemediationState
	now := time.Now()

	switch machineRemediation.Status.State {
	// initiating the reboot action
	case mrv1.RemediationStateStarted:
		// skip the reboot in case when the machine has power off state before the reboot action
		// it can mean that an user power off the machine by purpose
		if !bmh.Spec.Online {
			glog.V(4).Infof("Skip the remediation, machine %s has power off state before the remediation action", machine.Name)
			state = mrv1.RemediationStateSucceeded
			reason = "Skip the reboot, the machine power off by an user"
			newMachineRemediation.Status.EndTime = &metav1.Time{Time: now}
		} else {
			// power off the machine
			glog.V(4).Infof("Power off machine %s", machine.Name)
			newBmh.Spec.Online = false
			if err := bmr.client.Update(context.TODO(), newBmh); err != nil {
				return err
			}
			state = mrv1.RemediationStatePowerOff
			reason = "Starts the reboot process"
		}

	case mrv1.RemediationStatePowerOff:
		state = mrv1.RemediationStatePowerOn
		reason = "Reboot in progress"

		// host still has state on, we need to reconcile
		if bmh.Status.PoweredOn {
			glog.Warningf("machine %s still has power on state equal to true", machine.Name)
			return nil
		}

		// power on the machine
		glog.V(4).Infof("Power on machine %s", machine.Name)
		newBmh.Spec.Online = true
		if err := bmr.client.Update(context.TODO(), newBmh); err != nil {
			return err
		}

	case mrv1.RemediationStatePowerOn:
		// Node back to Ready under the cluster
		if NodeHasCondition(node, corev1.NodeReady, corev1.ConditionTrue) {
			glog.V(4).Infof("Remediation of machine %s succeeded", machine.Name)
			state = mrv1.RemediationStateSucceeded
			reason = "Reboot succeeded"
			newMachineRemediation.Status.EndTime = &metav1.Time{Time: now}
		}

	case mrv1.RemediationStateSucceeded:
		// remove reboot annotation when the reboot succeeded
		if _, ok := node.Annotations[annotationReboot]; ok {
			delete(node.Annotations, annotationReboot)
		}
		return bmr.client.Update(context.TODO(), node)

	case mrv1.RemediationStateFailed:
		// remove the unhealthy node from the cluster when the remediation failed
		// to free attached resources
		return bmr.client.Delete(context.TODO(), node)
	}

	// Reboot operation took more than defined timeout
	if machineRemediation.Status.StartTime.Time.Add(rebootDefaultTimeout * time.Minute).Before(now) {
		glog.Errorf("Remediation of machine %s failed on timeout", machine.Name)
		state = mrv1.RemediationStateFailed
		reason = "Reboot failed on timeout"
		newMachineRemediation.Status.EndTime = &metav1.Time{Time: now}
	}

	newMachineRemediation.Status.State = state
	newMachineRemediation.Status.Reason = reason
	glog.V(4).Infof("Update MachineRemediation %s status", machineRemediation.Name)
	return bmr.client.Update(context.TODO(), newMachineRemediation)
}

// GetBareMetalHostByMachine returns the bare metal host that linked to the machine
func GetBareMetalHostByMachine(c client.Client, machine *mapiv1.Machine) (*bmov1.BareMetalHost, error) {
	bmhKey, ok := machine.Annotations[annotationBareMetalHost]
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
