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

package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

const actionDiscover = "discover"

func NewAnsibleCommand() *cobra.Command {

	fence := &cobra.Command{
		Use:   "ansible",
		Short: "run ansible playbook",
		RunE:  ansible,
		Args:  cobra.ArbitraryArgs,
	}

	fence.Flags().String("agent-type", "", "Fencing agent type")
	fence.Flags().String("secret-path", "", "Path to the secret that contains fencing agent username and password")
	fence.Flags().StringP("action", "o", "", "Fencing action(status, reboot, off or on)")
	fence.Flags().String("playbook-path", "", "Path to ansible playbook to run(relevant only with --use-ansible flag)")
	return fence
}

func ansible(cmd *cobra.Command, args []string) (err error) {
	// Set power management agent type
	fenceCommand := "ansible-playbook"
	playbook, err := cmd.Flags().GetString("playbook-path")
	if err != nil {
		return err
	}

	extraVars := []string{}
	fenceAgentType, err := cmd.Flags().GetString("agent-type")
	if err != nil {
		return err
	}
	extraVars = append(extraVars, fmt.Sprintf("agent_type=%s", fenceAgentType))

	secretPath, err := cmd.Flags().GetString("secret-path")
	if err != nil {
		return err
	}
	// Set power management username
	username, err := ioutil.ReadFile(filepath.Join(secretPath, "username"))
	if err != nil {
		return err
	}
	extraVars = append(extraVars, fmt.Sprintf("username=%s", strings.Trim(string(username), "\n")))

	// Set power management password
	password, err := ioutil.ReadFile(filepath.Join(secretPath, "password"))
	if err != nil {
		return err
	}
	extraVars = append(extraVars, fmt.Sprintf("password=%s", strings.Trim(string(password), "\n")))

	// Set power management action
	action, err := cmd.Flags().GetString("action")
	if err != nil {
		return err
	}
	extraVars = append(extraVars, fmt.Sprintf("action=%s", action))

	// Set additional arguments
	options, err := cmd.Flags().GetString("options")
	if err == nil && options != "" {
		aOptions := []string{}
		optionsList := strings.Split(options, ",")
		for _, option := range optionsList {
			keyVal := strings.Split(option, "=")
			if len(keyVal) != 2 {
				return fmt.Errorf("incorrect option format, please use \"key1=value1,...,keyn=valuen\"")
			}

			arg := fmt.Sprintf("--%s=%s", keyVal[0], keyVal[1])
			if keyVal[1] == "" {
				arg = fmt.Sprintf("--%s", keyVal[0])
			}
			aOptions = append(aOptions, arg)
		}
		extraVars = append(extraVars, fmt.Sprintf("%s=\"%s\"", "options", strings.Join(aOptions, " ")))
	}

	glog.Infof("running ansible playbook %s with extra vars %s", playbook, extraVars)
	// Do not run command if dry-run is true
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	if dryRun {
		return nil
	}

	fenceArgs := []string{
		playbook,
		fmt.Sprintf("--extra-vars=%s", strings.Join(extraVars, " ")),
	}
	// Execute fence command
	_, stderr, rc := RunCommand(fenceCommand, fenceArgs...)
	if (action == actionDiscover && rc == 2) || rc == 0 {
		return nil
	}
	return fmt.Errorf(stderr)
}
