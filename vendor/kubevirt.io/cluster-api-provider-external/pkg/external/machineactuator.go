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
	//	"errors"
	"fmt"
	"time"

	"github.com/golang/glog"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterclient "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset"
	"sigs.k8s.io/cluster-api/pkg/errors"

	"kubevirt.io/cluster-api-provider-external/pkg/apis/providerconfig/v1alpha1"
	"kubevirt.io/cluster-api-provider-external/pkg/external/machinesetup"
)

const (
	checkEventAction  = "Status"
	createEventAction = "Create"
	deleteEventAction = "Delete"
	noEventAction     = ""
)

type ExternalClient struct {
	providerConfigCodec *v1alpha1.ExternalProviderConfigCodec
	machineSetupConfig  machinesetup.SetupConfig
	clusterclient       clusterclient.Interface
	kubeclient          kubernetes.Interface
	eventRecorder       record.EventRecorder
}

func NewMachineActuator(kubeclient kubernetes.Interface, clusterclient clusterclient.Interface, machineSetupConfigPath string) (*ExternalClient, error) {
	codec, err := v1alpha1.NewCodec()
	if err != nil {
		return nil, err
	}

	machineSetupConfig, err := machinesetup.NewSetupConfig(machineSetupConfigPath)
	if err != nil {
		return nil, err
	}

	glog.V(2).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclient.CoreV1().Events(metav1.NamespaceAll)})

	return &ExternalClient{
		providerConfigCodec: codec,
		kubeclient:          kubeclient,
		machineSetupConfig:  machineSetupConfig,
		clusterclient:       clusterclient,
		eventRecorder:       eventBroadcaster.NewRecorder(kubescheme.Scheme, corev1.EventSource{Component: "machine-actuator"}),
	}, nil
}

// Create actuator action powers on the machine
func (c *ExternalClient) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	// TODO: add some logic that will avoid running fencing command on the macine
	// that already has desired state
	glog.Infof("Power-on machine %s", machine.Name)
	if _, err := c.executeAction(createEventAction, cluster, machine, false); err != nil {
		glog.Errorf("Could not fence machine %s: %v", machine.Name, err)
		return c.handleMachineError(machine, errors.CreateMachine(
			"error power-on instance: %v", err), createEventAction)
	}

	c.eventRecorder.Eventf(machine, corev1.EventTypeNormal, "Created", "Power-on Machine %s", machine.Name)
	glog.Infof("Machine %s fencing operation succeeded", machine.Name)
	return nil
}

// Delete actuator action powers off the machine
func (c *ExternalClient) Delete(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	// TODO: add some logic that will avoid running fencing command on the macine
	// that already has desired state

	glog.Infof("Power-off machine %s", machine.Name)
	if _, err := c.executeAction(deleteEventAction, cluster, machine, true); err != nil {
		return c.handleMachineError(machine, errors.DeleteMachine(
			"error power-off instance: %v", err), deleteEventAction)
	}

	c.eventRecorder.Eventf(machine, corev1.EventTypeNormal, "Deleted", "Power-off Machine %s", machine.Name)
	return nil
}

// Update does not run any code
func (c *ExternalClient) Update(cluster *clusterv1.Cluster, goalMachine *clusterv1.Machine) error {
	glog.Infof("NOT IMPLEMENTED: update machine %s", goalMachine.Name)
	return nil
}

// Exists returns true, if machine is power-on
func (c *ExternalClient) Exists(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	glog.Infof("Checking if machine %s is power-on", machine.Name)
	nodeFenceState, err := c.executeAction(checkEventAction, cluster, machine, true)
	if err != nil {
		// TODO: we need to get output from the job
		glog.Infof("%s fence command on the machine %s failed: %v", checkEventAction, machine.Name, err)
		return false, err
	}

	glog.Infof("Machine %s has status power-on equals to %b", machine.Name, nodeFenceState)
	return nodeFenceState, nil
}

func (c *ExternalClient) executeAction(command string, cluster *clusterv1.Cluster, machine *clusterv1.Machine, doWait bool) (bool, error) {
	fencingConfig, err := c.chooseFencingConfig(machine)
	if err != nil {
		return false, err
	}

	fencingJob, err := createFencingJob(command, machine, fencingConfig)
	if err != nil {
		return false, err
	}

	backoff := wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   1.2,
		Steps:    5,
	}
	err = wait.ExponentialBackoff(backoff, func() (bool, error) {
		j, err := c.kubeclient.BatchV1().Jobs(machine.Namespace).Create(fencingJob)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			// Retry it as errors writing to the API server are common
			//util.JsonLogObject("Bad Job", job)
			return false, err
		}
		fencingJob = j
		glog.Infof("Job %v running for %s.", fencingJob.Name, machine.Name)
		return true, nil
	})

	if err != nil {
		return false, err
	}

	if doWait {
		nodeFenceState, err := c.waitForJob(fencingJob.Name, fencingJob.Namespace, -1)
		if err != nil {
			glog.Errorf("Job %v error: %v", fencingJob.Name, err)
			return false, err
		}
		return nodeFenceState, nil
	}

	glog.Infof("Job %v complete", fencingJob.Name)
	return true, nil
}

func (c *ExternalClient) waitForJob(jobName string, namespace string, retries int) (bool, error) {
	// TODO: Use informers to fetch resources
	job, err := c.kubeclient.BatchV1().Jobs(namespace).Get(jobName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	glog.Infof("Waiting %d times for job %v", retries, job.Name)

	for lpc := 0; lpc < retries || retries < 0; lpc++ {
		if len(job.Status.Conditions) > 0 {
			for _, condition := range job.Status.Conditions {
				if condition.Type == batchv1.JobFailed {
					return false, fmt.Errorf("Job %v failed: %v", job.Name, condition.Message)

				} else if condition.Type == batchv1.JobComplete {
					if job.Status.Succeeded > 0 {
						if job.DeletionTimestamp == nil {
							err := c.kubeclient.BatchV1().Jobs(namespace).Delete(job.Name, &metav1.DeleteOptions{})
							if err != nil {
								glog.Errorf("failed to delete succeeded job %s: %v", job.Name, err)
							}
						}
						return true, nil
					}
					return false, fmt.Errorf("Job %v failed: %v", job.Name, condition.Message)
				}
			}
		}
		time.Sleep(5 * time.Second)

		options := metav1.GetOptions{ResourceVersion: job.ObjectMeta.ResourceVersion}
		job, err = c.kubeclient.BatchV1().Jobs(namespace).Get(job.ObjectMeta.Name, options)
		if err != nil {
			return false, err
		}
	}

	return false, fmt.Errorf("Job %v in progress", job.Name)
}

func (c *ExternalClient) chooseFencingConfig(machine *clusterv1.Machine) (*v1alpha1.FencingConfig, error) {
	machineConfig, err := c.machineproviderconfig(machine.Spec.ProviderConfig)
	if err != nil {
		glog.Warningf("Could not unpack machine provider config for %s: %v", machine.Name, err)
		return nil, err
	}

	// Prefer fencing config defined as part of the Machine object over those
	// defined in the the Cluster
	if machineConfig.FencingConfig != nil {
		glog.Infof("Choose machine fencing configuration for machine %s", machine.Name)
		return machineConfig.FencingConfig, nil
	}

	machineParams := &machinesetup.MachineParams{
		Label: machineConfig.Label,
		Roles: machineConfig.Roles,
	}

	configMapMachine, err := c.machineSetupConfig.GetConfig(machineParams)
	if err != nil {
		return nil, err
	}

	if configMapMachine != nil {
		glog.Infof("Choose machine fencing configuration for machine %s from configMap", machine.Name)
		return configMapMachine.FencingConfig, nil
	}

	return nil, fmt.Errorf("No valid config for %v", machine.Name)
}

func (c *ExternalClient) machineproviderconfig(providerConfig clusterv1.ProviderConfig) (*v1alpha1.ExternalMachineProviderConfig, error) {
	var config v1alpha1.ExternalMachineProviderConfig
	err := c.providerConfigCodec.DecodeFromProviderConfig(providerConfig, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// If the ExternalClient has a client for updating Machine objects, this will set
// the appropriate reason/message on the Machine.Status. If not, such as during
// cluster installation, it will operate as a no-op. It also returns the
// original error for convenience, so callers can do "return handleMachineError(...)".
func (c *ExternalClient) handleMachineError(machine *clusterv1.Machine, err *errors.MachineError, eventAction string) error {
	if c.clusterclient != nil {
		reason := err.Reason
		message := err.Message
		machine.Status.ErrorReason = &reason
		machine.Status.ErrorMessage = &message
		c.clusterclient.ClusterV1alpha1().Machines(machine.Namespace).UpdateStatus(machine)
	}

	if eventAction != noEventAction {
		c.eventRecorder.Eventf(machine, corev1.EventTypeWarning, "Failed"+eventAction, "%v", err.Reason)
	}

	glog.Errorf("Machine error: %v", err.Message)
	return err
}
