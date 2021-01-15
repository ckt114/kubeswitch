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
	"github.com/trankchung/kubeswitch/kubeswitch"
)

// contextCmd represents the context command that presents a list
// of available contexts for user to pick from when no argument is passed.
// If argument is passed, which is the name of the context to switch to,
// then switch to that context without listing available contexts.
var contextCmd = &cobra.Command{
	Use:     "context",
	Short:   "List or set context",
	Aliases: []string{"ctx"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// Create an instance of Kubeswitch with passed in config if set.
		ks, err := kubeswitch.New()
		if err != nil {
			fail(err)
		}

		// Prompt user to select a context since no context is passed in.
		if len(args) < 1 {
			// Get string list of contexts.
			ctxs := *ks.ListContexts()

			// List context one per line without prompt. Use for shell completion.
			if viper.GetBool("noPrompt") {
				list(&ctxs)
			} else {
				// Prompt user to select context from a list.
				c, err := selectOption("context", ctxs)
				if err != nil {
					fail(err)
				}

				// Set to selected context picked from prompt.
				if err := ks.SetContext(c); err != nil {
					fail(err)
				}
			}
		} else {
			// Set to context provided as argument from command line.
			if err := ks.SetContext(args[0]); err != nil {
				fail(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
