package remediator

import (
	"context"

	mrv1 "github.com/openshift/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
)

// BareMetalRemediator implements Remediator interface for bare metal machines
type BareMetalRemediator struct {
}

// NewBareMetalRemediator returns new BareMetalRemediator object
func NewBareMetalRemediator() *BareMetalRemediator {
	return &BareMetalRemediator{}
}

// Recreate recreates the bare metal machine under the cluster
func (bmr *BareMetalRemediator) Recreate(ctx context.Context, machineRemediation *mrv1.MachineRemediation) error {
	return nil
}

// Reboot reboots the bare metal machine
func (bmr *BareMetalRemediator) Reboot(ctx context.Context, machineRemediation *mrv1.MachineRemediation) error {
	return nil
}
