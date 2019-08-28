// NOTE: Boilerplate only.  Ignore this file.

// Package v1alpha1 contains API Schema definitions for the healthchecking v1alpha1 API group
// +k8s:deepcopy-gen=package,register
// +groupName=machineremediation.kubevirt.io
package v1alpha1

import (
	"context"
	corev1 "k8s.io/api/core/v1"
        "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
        NamespaceOpenShiftMachineAPI = "openshift-machine-api"
        MasterNodeRoleLabel = "node-role.kubernetes.io/master"
)

func GetReplicaCount(c client.Client) int32 { 
        masterNodes := &corev1.NodeList{} 
        err := c.List( 
                context.TODO(), 
                masterNodes, 
                client.InNamespace(NamespaceOpenShiftMachineAPI), 
                client.MatchingLabels(map[string]string{MasterNodeRoleLabel: ""}), 
        ) 
        if err != nil { 
                return 1
        } 
        if len(masterNodes.Items) < 2 { 
                return 1
        } 
        return 2
} 

