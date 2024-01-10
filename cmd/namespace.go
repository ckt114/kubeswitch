/*
Copyright © 2020 Chung Tran <chung.k.tran@gmail.com>

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
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ckt114/kubeswitch/kubeswitch"
)

// namespaceCmd represents the namespace command that presents a list
// of available namespaces for user to pick from when no argument is passed.
// If argument is passed, which is the name of the namespace to switch to,
// then switch to that namespace without listing available namespaces.
var namespaceCmd = &cobra.Command{
	Use:     "namespace",
	Short:   "List and set namespace",
	Aliases: []string{"ns"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// Create an instance of Kubeswitch with config from default location.
		ks, err := kubeswitch.New()
		if err != nil {
			fail(err)
		}

		// Load namespaces for current context live from Kubernetes.
		if err := ks.LoadNamespaces(); err != nil {
			fail(err)
		}

		// Prompt user to select a namespace since no namespace is passed in.
		if len(args) < 1 {
			// Get a string list of namespaces.
			nss := *ks.ListNamespaces()

			// List namespaces one per line without prompt. Use for shell completion.
			if viper.GetBool("noPrompt") {
				list(&nss)
			} else {
				// Prompt user to select namespace from a list.
				n, err := selectOption("namespace", nss)
				if err != nil {
					fail(err)
				}

				// Set to selected namespace picked from prompt.
				if err := ks.SetNamespace(n); err != nil {
					fail(err)
				}
			}

		} else {
			// Set to namespace provided as argument from command line.
			if err := ks.SetNamespace(args[0]); err != nil {
				fail(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(namespaceCmd)
}
