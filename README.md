# Machine Remediation

## Remediation Flow

![Remediation Flow](docs/remediation-flow.png)

## Architecture

The machine remediation contains components to monitor and remediate unhealthy machines for different platforms, it works on top of [machine-api-operator](https://github.com/openshift/machine-api-operator) controllers.

It contains:

* [machine-remediation](docs/machine-remediation.md) controller
* [node-reboot](docs/node-reboot.md)

## How to deploy

You can check the [GitHub releases](https://github.com/kubevirt/machine-remediation/releases) to get latest `yaml` file, that includes CRD's, RBAC rules and deployment and apply it to your cluster.

```bash
kubectl apply -f https://github.com/kubevirt/machine-remediation/releases/download/v0.4.2/machine-remediation.v0.4.2.yaml
```

Once the deployment finishes, create a `MachineHealthCheck` object and be sure to give it the `healthchecking.openshift.io/strategy: reboot` annotation that instructs the Machine Healthcheck controller to delegate remediation to us.

An example `MachineHealthCheck` object that covers all nodes in the cluster is as follows:

```yaml
apiVersion: healthchecking.openshift.io/v1alpha1
kind: MachineHealthCheck
metadata:
 name: some-example
 namespace: openshift-machine-api
 annotations:
   healthchecking.openshift.io/strategy: reboot
spec:
 selector:
   matchLabels:
     kubernetes.io/os: linux
 unhealthyConditions:
 - type: Healthy
   status: Unknown
   timeout: 60s
```

## How to run e2e tests

You should have k8s or OpenShift environment with at least two worker nodes and run:

```bash
export KUBECONFIG=/dir/cluster/kubeconfig
make e2e-tests-run
```
