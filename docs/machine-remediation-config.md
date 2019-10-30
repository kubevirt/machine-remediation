# Machine Remediation Config

## Table of Contents

* [Machine Remediation Config](#machine-remediation-config)
  * [Status](#status)
  * [Table of Contents](#table-of-contents)
  * [Summary](#summary)
  * [Motivation](#motivation)
    * [Goals](#goals)
    * [Non-Goals](#non-goals)
  * [Proposal](#proposal)
    * [Implementation Details/Notes/Constraints](#implementation-detailsnotesconstraints)
      * [MachineRemediationConfig spec API](#machineremediationconfig-spec-api)
      * [MachineRemediationConfig status API](#machineremediationconfig-status-api)
  * [Design Details](#design-details)
    * [Work Items](#work-items)
    * [Alternatives](#alternatives)
    * [Dependencies](#dependencies)

## Summary

This document explains the implementation of the Machine Remediation Config CR and the motivation to create one.

## Motivation

Very often we need to have some global remediation configuration for the set of machines and we can not save it under the **MachineRemediation** object, because we create a new object for each remediation.

Some examples of such configuration:

1. The number of retries and intervals between remediation for the machine.
2. *Labels*, *annotation* and *taints* that we want to apply on the node that mapped to the machine after the remediation.

### Goals

1. To provide API that will allow a user to configure the flow of the remediation and apply some additional actions.
2. To provide information regarding the remediation flow to different controllers.
3. To provide controller to apply *labels*, *annotations* and *taints* on the node that mapped to the machine.

### Non-Goals

1. To provide from the beginning stable API, API can be changed in the future in the case when we will have requests from users.
2. To provide an additional logic under the bare metal remediation.

## Proposal

### Implementation Details/Notes/Constraints

The use will need to create the **MachineRemediationConfig** resource with the correct *labelSelector*, once it here a new controller that will apply the user-defined label on the node, will fetch data from it and apply *labels*, *annotations* and *taints* on the node that lacks it. The controller will decide if the node should be update by *labelSelector* -> *machine* -> *nodeRef*.

#### MachineRemediationConfig spec API

This proposal introduces a new API type: **MachineRemediationConfig**, that makes possible to create simple **MachineRemediationConfig** object with *labelSelector* and additional fields.

##### Example

```yaml
apiVersion: machineremediation.kubevirt.io/v1alpha1
kind: MachineRemediationConfig
metadata:
  name: mrc
  namespace: openshift-machine-api
spec:
  labelSelector: test
  nodeLabels:
    key1: value1
    key2: value2
  nodeAnnotations:
    key1: value1
    key2: value2
  nodeTaints:
  - key: key1
    value: value1
    effect: effect1
```

#### MachineRemediationConfig status API

**MachineRemediationStatus** currently does not contain any useful information, but in the future, it can contain the number of remediation retries for each machine.

### Risks and Mitigations

The most important question, how will we deliver it and will the OpenShift cloud team want it.

## Design Details

### Work Items

* Create new **MachineRemediationConfig** CRD.
* Write a new controller that will watch **MachineRemediationConfig** and **Node** objects.
* Deploy this controller as part of **MachineRemediation** controllers.

### Alternatives

It possible to intergrate *nodeLabels*, *nodeAnnotations* and *nodeTaints* to be bart of **MachineHealthCheck** resource, but I think it better to have separate controller for each functionallity.

### Dependencies

N/A
