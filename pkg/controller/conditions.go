/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package controller

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
)

type NodeConditionManager struct {
}

// GetNodeCondition returns node condition by type
func (d *NodeConditionManager) GetNodeCondition(node *corev1.Node, conditionType corev1.NodeConditionType) *corev1.NodeCondition {
	for _, cond := range node.Status.Conditions {
		if cond.Type == conditionType {
			return &cond
		}
	}
	return nil
}

func NewNodeConditionManager() *NodeConditionManager {
	return &NodeConditionManager{}
}

type RemediationConditions struct {
	Items []RemediationCondition `json:"items"`
}

type RemediationCondition struct {
	Name    string `json:"name"`
	Timeout string `json:"timeout"`
	Status  string `json:"status"`
}

func (d *NodeConditionManager) GetNodeRemediationConditions(node *corev1.Node, remediationConds *corev1.ConfigMap) ([]RemediationCondition, error) {
	data, ok := remediationConds.Data["conditions"]
	if !ok {
		return nil, fmt.Errorf("can not find \"conditions\" under configmap")
	}

	var remediationConditions RemediationConditions
	err := yaml.Unmarshal([]byte(data), &remediationConditions)
	if err != nil {
		glog.Errorf("failed to umarshal: %v", err)
		return nil, err
	}

	conditions := []RemediationCondition{}
	for _, c := range remediationConditions.Items {
		cond := d.GetNodeCondition(node, corev1.NodeConditionType(c.Name))
		if cond != nil && cond.Status == corev1.ConditionStatus(c.Status) {
			conditions = append(conditions, c)
		}
	}
	return conditions, nil
}
