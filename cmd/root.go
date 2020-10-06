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
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// Version will automatically be set to latest git tagged version.
var Version = "v0.0.0"

var (
	// pVersion is to print version number.
	pVersion bool

	// nPrompt disables select prompt for commands results.
	nPrompt bool

	// pSize is the select prompt size.
	pSize int

	// list prints result one per line.
	list = func(data *[]string) {
		fmt.Println(strings.Join(*data, "\n"))
	}

	// fail prints error message and exit.
	fail = func(err interface{}) {
		fmt.Println(err)
		os.Exit(1)
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubeswitch",
	Short: "Switch Kubernetes context or namespace",
	Run: func(cmd *cobra.Command, args []string) {
		if pVersion {
			fmt.Println(Version)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fail(err)
	}
}

func init() {
	cobra.OnInitialize()

	// Persistent flag that are available for all commands.
	rootCmd.PersistentFlags().BoolVarP(&nPrompt, "no-prompt", "P", false, "disable select prompt")
	rootCmd.PersistentFlags().IntVarP(&pSize, "prompt-size", "p", 10, "select prompt size")

	// Local flags only available to this command.
	rootCmd.Flags().BoolVarP(&pVersion, "version", "v", false, "print version")
}

func selectOption(kind string, data []string) (string, error) {
	// Function used for filtering result set.
	searcher := func(input string, index int) bool {
		name := data[index]
		return strings.Contains(name, input)
	}

	// Setup select prompt.
	prompt := promptui.Select{
		Label:             fmt.Sprintf("Select %s. / to search", kind),
		Items:             data,
		Size:              pSize,
		Searcher:          searcher,
		StartInSearchMode: false,
		HideHelp:          true,
		HideSelected:      false,
	}

	// Prompt user to select item from list.
	_, i, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return i, nil
}
