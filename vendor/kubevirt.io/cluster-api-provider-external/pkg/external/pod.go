/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package external

import (
	"fmt"
	"strings"

	"github.com/golang/glog"

	v1batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"kubevirt.io/cluster-api-provider-external/pkg/apis/providerconfig/v1alpha1"
)

const (
	secretsDir = "/etc/fencing/secrets"
)

func createFencingJob(action string, machine *clusterv1.Machine, fencingConfig *v1alpha1.FencingConfig) (*v1batch.Job, error) {
	// Create a Job with a container for each mechanism

	// TODO: Leverage podtemplates?
	containers := []v1.Container{}

	labels := map[string]string{"kubevirt.io/cluster-api-provider-external": action}

	container := fencingConfig.Container.DeepCopy()

	fencingArgs, err := getFencingCommand(container, fencingConfig, action, machine.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get fencing command: %v", err)
	}
	container.Args = fencingArgs
	
	volumes := processSecret(fencingConfig, container)

	// Attach additional volumes to the job pod
	for _, v := range fencingConfig.Volumes {
		volumes = append(volumes, v)
	}

	// Add the container to the PodSpec
	containers = append(containers, *container)

	timeout := int64(30) // TODO: Make this configurable
	numContainers := int32(1)

	// Parallel Jobs with a fixed completion count
	// - https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	return &v1batch.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%v-job-%v-", machine.ObjectMeta.Name, strings.ToLower(action)),
			Namespace:    machine.Namespace,
			Labels: labels,
		},
		Spec: v1batch.JobSpec{
			BackoffLimit:          fencingConfig.Retries,
			Parallelism:           &numContainers,
			Completions:           &numContainers,
			ActiveDeadlineSeconds: &timeout,
			// https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/#clean-up-finished-jobs-automatically
			// TTLSecondsAfterFinished: 100,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers:    containers,
					RestartPolicy: v1.RestartPolicyOnFailure,
					Volumes:       volumes,
				},
			},
		},
	}, nil
}

func processSecret(fencingConfig *v1alpha1.FencingConfig, c *v1.Container) []v1.Volume {
	// Create volumes for any sensitive parameters that are stored as k8s secrets
	volumeName := "fencing-secret"
	volumes := []v1.Volume{
		{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{SecretName:fencingConfig.Secret},
			},
		},
	}

	// Mount the secrets into the container so they can be easily retrieved
	c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: secretsDir,
	})
	
	// Append secreds to container arguments
	c.Args = append(c.Args, fmt.Sprintf("--secret-path=%s", secretsDir))

	return volumes
}

func getFencingCommand(c *v1.Container, fencingConfig *v1alpha1.FencingConfig, action string, target string) ([]string, error) {
	fencingCommand := []string{}
	if c.Args != nil {
		fencingCommand = c.Args
	}

	switch action {
	case createEventAction:
		fencingCommand = append(fencingCommand, fencingConfig.CreateArgs...)
	case deleteEventAction:
		fencingCommand = append(fencingCommand, fencingConfig.DeleteArgs...)
	case checkEventAction:
		fencingCommand = append(fencingCommand, fencingConfig.CheckArgs...)
	default:
		return fencingCommand, fmt.Errorf("unsupported fencing action for the machine %s", target)
	}

	if fencingConfig.DynamicConfig != nil {
		options := []string{}
		for _, dc := range fencingConfig.DynamicConfig {
			if value, ok := dc.Lookup(target); ok {
				options = append(options, fmt.Sprintf("%s=%s", dc.Field, value))
			} else {
				glog.Warningf("no value of '%s' found for '%s'", dc.Field, target)
			}
		}
		fencingCommand = append(fencingCommand, fmt.Sprintf("--options=%s", strings.Join(options, ",")))
	}

	glog.Infof("%s %v fencingCommand: %v", fencingConfig.Container.Name, action, fencingCommand)
	return fencingCommand, nil
}

