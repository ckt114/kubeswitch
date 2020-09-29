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

	"github.com/spf13/cobra"
	"github.com/trankchung/kubeswitch/kubeswitch"
)

var (
	// days is number of days to retain copied session files.
	days int
)

// purgeCmd represents the purge command that purges temporary session files.
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge temporary session files",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("purging temporary session files older than %d day(s) ...\n", days)
		kubeswitch.Purge(days)
		fmt.Println("done")
	},
}

func init() {
	rootCmd.AddCommand(purgeCmd)

	// Local flags only available to this command.
	purgeCmd.Flags().IntVarP(&days, "days", "d", 2, "days to retain")
}
