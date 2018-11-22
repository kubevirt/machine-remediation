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

	volumeMap := map[string]v1.Volume{}
	containers := []v1.Container{}

	labels := map[string]string{"kubevirt.io/cluster-api-external-provider": action}

	container := fencingConfig.Container.DeepCopy()

	fencingCommand, err := getFencingCommand(container, fencingConfig, action, machine.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get fencing command: %v", err)
	}
	container.Args = fencingCommand

	env, err := getContainerEnv(fencingConfig, action, machine.Name, secretsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to set environment variables: %v", err)
	}
	container.Env = env

	for _, v := range processSecrets(fencingConfig, container) {
		if _, ok := volumeMap[v.Name]; !ok {
			volumeMap[v.Name] = v
		}
	}
	// Add the container to the PodSpec
	containers = append(containers, *container)

	volumes := []v1.Volume{}
	if fencingConfig.Volumes != nil {
		volumes = fencingConfig.Volumes
	}

	for _, v := range volumeMap {
		volumes = append(volumes, v)
	}

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

func volumeNameMap(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return 'a' + (r - 'A')
	case r >= 'a' && r <= 'z':
		return r
	case r >= '0' && r <= '9':
		return r
	default:
		return '-'
	}
}

func processSecrets(fencingConfig *v1alpha1.FencingConfig, c *v1.Container) []v1.Volume {
	volumes := []v1.Volume{}
	for key, s := range fencingConfig.Secrets {

		// volumeName must contain only a-z, 0-9, and -
		volumeName := strings.Map(volumeNameMap, fmt.Sprintf("secret-%s", key))
		mount := fmt.Sprintf("%s/%s-%s", secretsDir, s, key)
		data := fmt.Sprintf("%s/%s", mount, key)

		// Create volumes for any sensitive parameters that are stored as k8s secrets
		volumes = append(volumes, v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: s,
				},
			},
		})

		// Relies on an ENTRYPOINT that looks for SECRETPATH_field=/path/to/file and adds: --field=$(cat /path/to/file) to the command line
		c.Env = append(c.Env, v1.EnvVar{
			Name:  fmt.Sprintf("SECRETPATH_%s", key),
			Value: data,
		})

		// Mount the secrets into the container so they can be easily retrieved
		c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
			Name:      volumeName,
			ReadOnly:  true,
			MountPath: mount,
		})
	}
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

	if fencingConfig.ArgumentFormat == "env" {

		if len(fencingConfig.PassTargetAs) == 0 {
			// No other way to pass it in, just append to the existing fencingCommand
			fencingCommand = append(fencingCommand, target)
		}

	} else if fencingConfig.ArgumentFormat == "cli" {
		for name, value := range fencingConfig.Config {
			fencingCommand = append(fencingCommand, fmt.Sprintf("--%s", name))
			fencingCommand = append(fencingCommand, value)
		}

		for _, dc := range fencingConfig.DynamicConfig {
			fencingCommand = append(fencingCommand, fmt.Sprintf("--%s", dc.Field))
			if value, ok := dc.Lookup(target); ok {
				fencingCommand = append(fencingCommand, value)
			} else {
				return fencingCommand, fmt.Errorf("no value of '%s' found for '%s'", dc.Field, target)
			}
		}

		if len(fencingConfig.PassActionAs) > 0 {
			fencingCommand = append(fencingCommand, fmt.Sprintf("--%s", fencingConfig.PassActionAs))
			fencingCommand = append(fencingCommand, action)
		}

		if len(fencingConfig.PassTargetAs) > 0 {
			fencingCommand = append(fencingCommand, fmt.Sprintf("--%s", fencingConfig.PassTargetAs))
			fencingCommand = append(fencingCommand, target)
		}

	} else {
		return fencingCommand, fmt.Errorf("argumentFormat %s not supported", fencingConfig.ArgumentFormat)
	}

	glog.Infof("%s %v fencingCommand: %v", fencingConfig.Container.Name, action, fencingCommand)
	return fencingCommand, nil
}

func getContainerEnv(fencingConfig *v1alpha1.FencingConfig, action string, target string, secretsDir string) ([]v1.EnvVar, error) {
	env := []v1.EnvVar{
		{
			Name:  "ARG_FORMAT",
			Value: fencingConfig.ArgumentFormat,
		},
	}

	for _, val := range fencingConfig.Container.Env {
		env = append(env, val)
	}

	if fencingConfig.ArgumentFormat == "cli" {
		return env, nil
	}

	if fencingConfig.ArgumentFormat == "env" {
		glog.Infof("Adding env vars")
		for name, value := range fencingConfig.Config {
			glog.Infof("Adding %v=%v", name, value)
			env = append(env, v1.EnvVar{
				Name:  name,
				Value: value,
			})
		}

		glog.Infof("Adding dynamic env vars: %v", fencingConfig.DynamicConfig)
		for _, dc := range fencingConfig.DynamicConfig {
			if value, ok := dc.Lookup(target); ok {
				glog.Infof("Adding %v=%v (dynamic)", dc.Field, value)
				env = append(env, v1.EnvVar{
					Name:  dc.Field,
					Value: value,
				})
			} else {
				glog.Errorf("not adding %v (dynamic)", dc.Field)
				return nil, fmt.Errorf("no value of '%s' found for '%s'", dc.Field, target)
			}
		}

		if len(fencingConfig.PassTargetAs) > 0 {
			env = append(env, v1.EnvVar{
				Name:  fencingConfig.PassTargetAs,
				Value: target,
			})
		}

		if len(fencingConfig.PassActionAs) > 0 {
			env = append(env, v1.EnvVar{
				Name:  fencingConfig.PassActionAs,
				Value: action,
			})
		}

		return env, nil
	}
	return env, fmt.Errorf("argumentFormat %s not supported", fencingConfig.ArgumentFormat)
}
