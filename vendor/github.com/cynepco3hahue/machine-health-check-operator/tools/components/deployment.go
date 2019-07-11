package components

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// NewOperatorDeployment returns deployment object that represents machine-health-check-operator
func NewOperatorDeployment(namespace string, repository string, version string, pullPolicy corev1.PullPolicy, verbosity string) (*appsv1.Deployment, error) {
	name := "machine-health-check-operator"
	image := fmt.Sprintf("%s/%s:%s", repository, name, version)
	labels := map[string]string{"k8s-app": "machine-health-check-operator"}
	tolerations := []corev1.Toleration{
		{
			Key:    "node-role.kubernetes.io/master",
			Effect: corev1.TaintEffectNoSchedule,
		},
		{
			Key:      "CriticalAddonsOnly",
			Operator: corev1.TolerationOpExists,
		},
		{
			Key:               "node.kubernetes.io/not-ready",
			Effect:            corev1.TaintEffectNoExecute,
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: pointer.Int64Ptr(120),
		},
		{
			Key:               "node.kubernetes.io/unreachable",
			Effect:            corev1.TaintEffectNoExecute,
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: pointer.Int64Ptr(120),
		},
	}
	resources := corev1.ResourceRequirements{
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceMemory: resource.MustParse("50Mi"),
			corev1.ResourceCPU:    resource.MustParse("10m"),
		},
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "machine-api-operator",
					PriorityClassName:  "system-node-critical",
					NodeSelector:       map[string]string{"node-role.kubernetes.io/master": ""},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: pointer.BoolPtr(true),
						RunAsUser:    pointer.Int64Ptr(65534),
					},
					Tolerations:   tolerations,
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           image,
							ImagePullPolicy: pullPolicy,
							Command:         []string{"/usr/bin/machine-health-check-operator"},
							Args: []string{
								"start",
								"--alsologtostderr",
								"-v",
								verbosity,
							},
							Env: []corev1.EnvVar{
								{
									Name: "COMPONENT_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							Resources: resources,
						},
					},
				},
			},
		},
	}

	return deployment, nil
}
