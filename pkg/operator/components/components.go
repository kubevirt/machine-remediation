package components

const (
	// ComponentMachineDisruptionBudget contains name for MachineDisruptionBudget component
	ComponentMachineDisruptionBudget = "machine-disruption-budget"
	// ComponentMachineHealthCheck contains name for MachineHealthCheck component
	ComponentMachineHealthCheck = "machine-health-check"
	// ComponentMachineRemediation contains name for MachineRemediation component
	ComponentMachineRemediation = "machine-remediation"
)

var (
	// Components contains names of all componenets that the operator should deploy
	Components = []string{
		ComponentMachineDisruptionBudget,
		ComponentMachineHealthCheck,
		ComponentMachineRemediation,
	}
)
