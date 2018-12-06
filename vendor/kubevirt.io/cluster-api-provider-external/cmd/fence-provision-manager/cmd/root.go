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
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewRootCommand() *cobra.Command {

	root := &cobra.Command{
		Use:   "fence-provision-manager",
		Short: "fence-provision-manager can execute fencing and provisioning actions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStderr(), cmd.UsageString())
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().Bool("dry-run", false, "Dry run of the command, it will only log the command, but will not execute it")
	root.PersistentFlags().String("options", "", "Additional options passed to the command(key=value,...)")

	root.AddCommand(
		NewFenceCommand(),
		NewAnsibleCommand(),
	)

	return root

}

func Execute() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	flag.Set("logtostderr", "true")
	flag.Parse()

	if err := NewRootCommand().Execute(); err != nil {
		glog.Error(err)
		os.Exit(1)
	}
}
