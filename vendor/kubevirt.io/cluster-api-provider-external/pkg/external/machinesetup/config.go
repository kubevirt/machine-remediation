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

package machinesetup

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/ghodss/yaml"
	"kubevirt.io/cluster-api-provider-external/pkg/apis/providerconfig/v1alpha1"
)

// SetupConfig define interface to work with machine setup config
type SetupConfig interface {
	GetConfig(params *MachineParams) (*Config, error)
}

// SetupConfigImpl holds the path to the machine setup configs yaml file
type SetupConfigImpl struct {
	machineSetupPath string
}

type MachineConfigList struct {
	Items []MachineConfig `json:"items"`
}

// MachineConfig specify mapping between machine params to specific configuration
type MachineConfig struct {
	MachineParams []MachineParams `json:"machineParams,omitempty"`
	Config        *Config         `json:"config,omitempty"`
}

type Config struct {
	StartupScript string                  `json:"startupScript,omitempty"`
	FencingConfig *v1alpha1.FencingConfig `json:"fencingConfig,omitempty"`
}

type MachineParams struct {
	Label string                 `json:"label,omitempty"`
	Roles []v1alpha1.MachineRole `json:"roles,omitempty"`
}

func NewSetupConfig(path string) (*SetupConfigImpl, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &SetupConfigImpl{machineSetupPath: path}, nil
}

func (s *SetupConfigImpl) GetConfig(params *MachineParams) (*Config, error) {
	machineConfig, err := s.matchMachineConfig(params)
	if err != nil {
		return nil, err
	}
	return machineConfig.Config, nil
}

func (s *SetupConfigImpl) matchMachineConfig(params *MachineParams) (*MachineConfig, error) {
	machineConfigs, err := s.parseMachineSetupYaml()
	if err != nil {
		return nil, err
	}

	matchingConfigs := make([]MachineConfig, 0)
	for _, conf := range machineConfigs.Items {
		for _, validParams := range conf.MachineParams {
			if params.Label != validParams.Label {
				continue
			}

			validRoles := rolesToMap(validParams.Roles)
			paramRoles := rolesToMap(params.Roles)
			if !reflect.DeepEqual(paramRoles, validRoles) {
				continue
			}

			matchingConfigs = append(matchingConfigs, conf)
		}
	}

	if len(matchingConfigs) == 1 {
		return &matchingConfigs[0], nil
	} else if len(matchingConfigs) == 0 {
		return nil, fmt.Errorf("could not find a matching machine setup config for params %+v", params)
	} else {
		return nil, fmt.Errorf("found multiple matching machine setup configs for params %+v", params)
	}
}

func rolesToMap(roles []v1alpha1.MachineRole) map[v1alpha1.MachineRole]int {
	rolesMap := map[v1alpha1.MachineRole]int{}
	for _, role := range roles {
		rolesMap[role] = rolesMap[role] + 1
	}
	return rolesMap
}

func (s *SetupConfigImpl) parseMachineSetupYaml() (*MachineConfigList, error) {
	f, err := os.Open(s.machineSetupPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	machineConfigs := &MachineConfigList{}
	err = yaml.Unmarshal(bytes, machineConfigs)
	if err != nil {
		return nil, err
	}
	return machineConfigs, nil
}
