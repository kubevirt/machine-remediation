package components

const (
	// ComponentMachineRemediation contains name for MachineRemediation component
	ComponentMachineRemediation = "machine-remediation"
	// ComponentMachineRemediationOperator contains name for MachineRemediationOperator component
	ComponentMachineRemediationOperator = "machine-remediation-operator"
)

var (
	// Components contains names of all componenets that the operator should deploy
	Components = []string{
		ComponentMachineRemediation,
	}
)

const (
	// EnvVarOperatorVersion contains the name of operator version environment variable
	EnvVarOperatorVersion = "OPERATOR_VERSION"
)

const (
	// CRDMachineRemediation contains the kind of the MachineRemediation CRD
	CRDMachineRemediation = "machineremediations"
)

var (
	// CRDS contains names of all CRD's that the operator should deploy
	CRDS = []string{
		CRDMachineRemediation,
	}
)
