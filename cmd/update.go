/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/talos-systems/bldr/internal/pkg/update"
)

// updateCmd represents the update command.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "TODO",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			log.Print(update.Latest(arg))
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
