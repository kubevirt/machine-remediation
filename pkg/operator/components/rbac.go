package components

import (
	"fmt"

	"kubevirt.io/machine-remediation-operator/pkg/consts"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// Rules contains rules for all machine remediation components
	Rules = map[string][]rbacv1.PolicyRule{
		ComponentMachineDisruptionBudget: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"machine.openshift.io",
				},
				Resources: []string{
					rbacv1.ResourceAll,
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"machineremediation.kubevirt.io",
				},
				Resources: []string{
					"machinedisruptionbudgets",
				},
				Verbs: []string{
					"get",
					"list",
					"update/status",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"configmaps",
					"nodes",
				},
				Verbs: []string{
					"get",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"events",
				},
				Verbs: []string{
					"create",
					"list",
					"watch",
					"patch",
				},
			},
		},
		ComponentMachineHealthCheck: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"machine.openshift.io",
				},
				Resources: []string{
					"machine",
				},
				Verbs: []string{
					"delete",
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"machineremediation.kubevirt.io",
				},
				Resources: []string{
					"machinedisruptionbudgets",
					"machinehealthcheck",
				},
				Verbs: []string{
					"get",
					"list",
					"update/status",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"machineremediation.kubevirt.io",
				},
				Resources: []string{
					"machineremediation",
				},
				Verbs: []string{
					"create",
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"configmaps",
				},
				Verbs: []string{
					"get",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"nodes",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
		},
		ComponentMachineRemediation: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"machine.openshift.io",
				},
				Resources: []string{
					"machine",
				},
				Verbs: []string{
					"get",
					"list",
				},
			},
			{
				APIGroups: []string{
					"machineremediation.kubevirt.io",
				},
				Resources: []string{
					"machineremediations",
				},
				Verbs: []string{
					"delete",
					"get",
					"list",
					"update",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"nodes",
				},
				Verbs: []string{
					"delete",
					"get",
					"list",
				},
			},
			{
				APIGroups: []string{
					"metal3.io",
				},
				Resources: []string{
					"baremetalhosts",
				},
				Verbs: []string{
					"get",
					"list",
					"update",
				},
			},
		},
	}
)

// NewServiceAccount returns new ServiceAccount object
func NewServiceAccount(name string, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      fmt.Sprintf("%s-service-account", name),
			Labels: map[string]string{
				consts.LabelKubeVirt: "",
			},
		},
	}
}

// NewClusterRole returns new ClusterRole object
func NewClusterRole(name string, rules []rbacv1.PolicyRule) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-cluster-role", name),
			Labels: map[string]string{
				consts.LabelKubeVirt: "",
			},
		},
		Rules: rules,
	}
}

// NewClusterRoleBinding returns new ClusterRoleBinding object
func NewClusterRoleBinding(name string, namespace string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-cluster-role-binding", name),
			Labels: map[string]string{
				consts.LabelKubeVirt: "",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     fmt.Sprintf("%s-cluster-role", name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: namespace,
				Name:      fmt.Sprintf("%s-service-account", name),
			},
		},
	}
}
