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

const actionStatus = "status"

func NewFenceCommand() *cobra.Command {

	fence := &cobra.Command{
		Use:   "fence",
		Short: "run fencing command on the host",
		RunE:  fence,
		Args:  cobra.ArbitraryArgs,
	}

	fence.Flags().String("agent-type", "", "Fencing agent type")
	fence.Flags().String("secret-path", "", "Path to the secret that contains fencing agent username and password")
	fence.Flags().StringP("action", "o", "", "Fencing action(status, reboot, off or on)")
	return fence
}

func fence(cmd *cobra.Command, args []string) (err error) {
	// Set power management agent type
	fenceAgentType, err := cmd.Flags().GetString("agent-type")
	if err != nil {
		return err
	}
	fenceCommand := filepath.Join("/sbin", fmt.Sprintf("fence_%s", fenceAgentType))

	fenceArgs := []string{}

	secretPath, err := cmd.Flags().GetString("secret-path")
	if err != nil {
		return err
	}
	// Set power management username
	username, err := ioutil.ReadFile(filepath.Join(secretPath, "username"))
	if err != nil {
		return err
	}
	fenceArgs = append(fenceArgs, fmt.Sprintf("--username=%s", strings.Trim(string(username), "\n")))

	// Set power management password
	password, err := ioutil.ReadFile(filepath.Join(secretPath, "password"))
	if err != nil {
		return err
	}
	fenceArgs = append(fenceArgs, fmt.Sprintf("--password=%s", strings.Trim(string(password), "\n")))

	// Set power management action
	action, err := cmd.Flags().GetString("action")
	if err != nil {
		return err
	}
	fenceArgs = append(fenceArgs, fmt.Sprintf("--action=%s", action))

	// Set additional arguments
	options, err := cmd.Flags().GetString("options")
	if err == nil && options != "" {
		optionList := strings.Split(options, ",")
		for _, option := range optionList {
			keyVal := strings.Split(option, "=")
			if len(keyVal) != 2 {
				return fmt.Errorf("incorrect option format, please use \"key1=value1,...,keyn=valuen\"")
			}
			
			arg := fmt.Sprintf("--%s=%s", keyVal[0], keyVal[1])
			if keyVal[1] == "" {
				arg = fmt.Sprintf("--%s", keyVal[0])
			}
			fenceArgs = append(fenceArgs, arg)
		}
	}

	glog.Infof("running fence command %s with arguments %s", fenceCommand, fenceArgs)
	// Do not run command if dry-run is true
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	if dryRun {
		return nil
	}

	// Execute fence command
	_, stderr, rc := RunCommand(fenceCommand, fenceArgs...)
	if (action == actionStatus && rc == 2) || rc == 0 {
		return nil
	}
	return fmt.Errorf(stderr)
}
