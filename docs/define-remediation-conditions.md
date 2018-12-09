# Remediation conditions

The remediation controller will start the work only when node has one of conditions specified under `remediation-conditions` config map.

### Defaults

By default we provide only node `Ready` condition under the config map with the timeout of `60s`, it means that if node has `Ready` condition in the state `Unknown` more than `60s`, the controller will start remediate process.

### User-definied conditions

You can easily add additional conditions under the config map or edit already existing conditions:

```
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    noderecovery.kubevirt.io: ""
  name: remediation-conditions
  namespace: noderecovery
data:
  conditions: |
    items:
    - name: Ready     # Condition name
      timeout: 120s   # Condition timeout
      status: Unknown # Condition status
    - name: Ready
      timeout: 60s
      status: False
```
