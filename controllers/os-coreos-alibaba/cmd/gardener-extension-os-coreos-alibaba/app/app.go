// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"fmt"
	"github.com/gardener/gardener-extensions/controllers/os-coreos-alibaba/pkg/coreos-alibaba"
	"github.com/gardener/gardener-extensions/pkg/controller"
	"github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig"
	"github.com/spf13/cobra"
	"os"
)

const Name = "os-coreos-alibaba"

// ActuatorFactory is the factory to create a CoreOS Alibaba Actuator.
func ActuatorFactory(args *operatingsystemconfig.ActuatorArgs) (operatingsystemconfig.Actuator, error) {
	return coreos.NewActuator(args.Log), nil
}

// NewControllerCommand creates a new command for running a CoreOS Alibaba controller.
func NewControllerCommand() *cobra.Command {
	opts := operatingsystemconfig.NewCommandOptions(Name, coreos.Type, ActuatorFactory)
	opts.Manager.LeaderElection = true
	opts.Manager.LeaderElectionNamespace = os.Getenv("LEADER_ELECTION_NAMESPACE")

	cmd := &cobra.Command{
		Use: "os-coreos-controller-manager",

		Run: func(cmd *cobra.Command, args []string) {
			c, err := opts.Config()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			if err := operatingsystemconfig.Run(controller.SetupSignalHandlerContext(), c.Complete()); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	fs := cmd.Flags()
	for _, f := range opts.Flags().FlagSets {
		fs.AddFlagSet(f)
	}

	return cmd
}
