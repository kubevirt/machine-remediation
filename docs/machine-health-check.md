# Machine Health Check

This controller responsible to monitor nodes and create the `MachineRemediation` object when the node has unhealthy condition, when unhealthy condition can be definied by a user.

## How to define `MachineHealthCheck`

```yaml
apiVersion: machineremediation.kubevirt.io/v1alpha1
kind: MachineHealthCheck
metadata:
  name: workers
  namespace: openshift-machine-api
spec:
  selector:
    matchLabels:
      machine.openshift.io/cluster-api-machine-role: worker
      machine.openshift.io/cluster-api-machine-type: worker
```

This health check will start to monitor all nodes with labels under the `matchLabels`.

## How to monitor custom node conditions

By default, the machine health check controller recognize only `NotReady` condition and will remove unhealthy machine after 5 minutes. If you want to customize unhealthy conditions you can create `node-unhealthy-conditions` config map, for example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: node-unhealthy-conditions
  namespace: openshift-machine-api
data:
  conditions: |
    items:
    - name: NetworkUnavailable
      timeout: 5s
      status: True
```