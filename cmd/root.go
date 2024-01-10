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

	"path/filepath"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ckt114/kubeswitch/kubeswitch"
)

const (
	defaultCfg = "$HOME/.kubeswitch.yaml"
)

// Version will automatically be set to latest git tagged version.
var Version = "v0.0.0"

var (
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
		if viper.GetBool("version") {
			fmt.Println(Version)
		} else if viper.GetBool("debug") {
			fmt.Println("KUBECONFIG:", os.Getenv(kubeswitch.EnvVarConfig))
			fmt.Println("Kubeswitch config:", viper.ConfigFileUsed())
			fmt.Printf("Config Values: %+v\n", viper.AllSettings())
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
	cobra.OnInitialize(initConfig)

	// Persistent flag that are available for all commands.
	rootCmd.PersistentFlags().StringP("config", "c", defaultCfg, "kubeswitch config (KUBESWITCH_CONFIG)")
	rootCmd.PersistentFlags().BoolP("no-config", "C", false, "don't use kubeswitch config (KUBESWITCH_NOCONFIG)")
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "kubernetes config to read (KUBESWITCH_KUBECONFIG)")
	rootCmd.PersistentFlags().IntP("prompt-size", "p", 10, "selection prompt size (KUBESWITCH_PROMPTSIZE)")
	rootCmd.PersistentFlags().BoolP("no-prompt", "P", false, "disable selection prompt (KUBESWITCH_NOPROMPT)")

	// Local flags only available to this command.
	rootCmd.Flags().BoolP("version", "v", false, "print version")
	rootCmd.Flags().BoolP("debug", "d", false, "print debug info")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Configure environment variable reading.
	viper.SetEnvPrefix("KUBESWITCH")
	viper.AutomaticEnv()

	viper.BindPFlag("config", rootCmd.Flags().Lookup("config"))
	viper.BindPFlag("noConfig", rootCmd.Flags().Lookup("no-config"))
	viper.BindPFlag("kubeConfig", rootCmd.Flags().Lookup("kubeconfig"))
	viper.BindPFlag("promptSize", rootCmd.Flags().Lookup("prompt-size"))
	viper.BindPFlag("noPrompt", rootCmd.Flags().Lookup("no-prompt"))

	viper.BindPFlag("version", rootCmd.Flags().Lookup("version"))
	viper.BindPFlag("debug", rootCmd.Flags().Lookup("debug"))

	// Only read Kubeswitch config file if `noConfig` is false.
	if !viper.GetBool("noConfig") {
		cfg, _ := homedir.Expand(os.ExpandEnv(viper.GetString("config")))
		viper.SetConfigFile(cfg)

		// Read Kubeswitch config if file exists.
		if _, err := os.Stat(viper.ConfigFileUsed()); err == nil {
			if err := viper.ReadInConfig(); err != nil {
				fail(fmt.Sprintln(viper.ConfigFileUsed(), ":", err))
			}
		} else {
			fmt.Printf("WARN: Config file \"%s\" not exists\n", viper.ConfigFileUsed())
		}
	}

	// Setup KUBECONFIG from flags, env vars, and config file.
	if err := setupKubeEnvVar(); err != nil {
		fail(err)
	}
}

// setupKubeEnvVar finds all the Kubernetes configs defined in Kubeswitch config file
// and construct into colon-separated list and set KUBECONFIG env var to that list.
// This is so that clientcmd can read multiple config at once.
func setupKubeEnvVar() error {
	if !kubeswitch.IsActive() {
		var configs []string

		// Add kubeConfig into list of configs.
		cfg, err := homedir.Expand(os.ExpandEnv(viper.GetString("kubeConfig")))
		if err != nil {
			return err
		}
		configs = append(configs, cfg)

		// Add KUBECONFIG into list of configs if defined.
		kConfig, err := homedir.Expand(os.ExpandEnv(os.Getenv(kubeswitch.EnvVarConfig)))
		if err != nil {
			return err
		}
		configs = append(configs, kConfig)

		// Get list of files matching patterns in `configs` key.
		for _, path := range viper.GetStringSlice("configs") {
			absPath, _ := homedir.Expand(os.ExpandEnv(path))
			files, _ := filepath.Glob(absPath)
			configs = append(configs, files...)
		}

		// Remove duplicate config paths from `configs`.
		configs = removeDuplicates(configs)

		// Set KUBECONFIG to list of configs separated by colon.
		if err := os.Setenv(kubeswitch.EnvVarConfig, strings.Join(configs, ":")); err != nil {
			return err
		}
	}

	return nil
}

func removeDuplicates(s []string) []string {
	items := map[string]bool{}

	// Create a map of all unique items.
	for v := range s {
		items[s[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range items {
		result = append(result, key)
	}
	return result
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
		Size:              viper.GetInt("promptSize"),
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
