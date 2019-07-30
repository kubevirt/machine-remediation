# Machine Remediation Operator

## Remediation Flow

![Remediation Flow](docs/remediation-flow.png)

## Architecture

The machine remediation operator deploys fencing controllers to provide remediation solution for different platforms, it works on top of cluster-api controllers.

It should deploy three controllers:

- [machine-health-check](docs/machine-health-check.md) controller
- [machine-disruption-budget](docs/machine-disruption-budget.md) controller
- [machine-remediation](docs/machine-remediation.md) controller

## How to deploy

You can check the [GitHub releases](https://github.com/kubevirt/machine-remediation-operator/releases) to get latest `yaml` file, that includes CRD's, RBAC rules and operator deployment and apply it to your cluster.

```bash
kubectl apply -f https://github.com/kubevirt/machine-remediation-operator/releases/download/v0.3.3/machine-remediation-operator.yaml
kubectl apply -f https://github.com/kubevirt/machine-remediation-operator/releases/download/v0.3.3/machine-remediation-operator-cr.yaml
```

After just wait until the operator will deploy all components.

## How to run e2e tests

You should have k8s or OpenShift environment with at least two worker nodes and run:

```bash
export KUBECONFIG=/dir/cluster/kubeconfig
make e2e-tests-run
```
