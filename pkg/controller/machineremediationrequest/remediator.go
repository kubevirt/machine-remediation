package machineremediationrequest

import (
	"context"

	mapiv1 "sigs.k8s.io/cluster-api/pkg/apis/machine/v1beta1"
)

// Remediator apply machine remediation strategy under a specific infrastructure.
type Remediator interface {
	// Reboot the machine.
	Reboot(context.Context, *mapiv1.Machine) error
	// Recreate the machine.
	Recreate(context.Context, *mapiv1.Machine) error
}
