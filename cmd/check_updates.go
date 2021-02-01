/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/talos-systems/bldr/internal/pkg/solver"
	"github.com/talos-systems/bldr/internal/pkg/update"
)

// checkUpdatesCmd represents the check-updates command.
var checkUpdatesCmd = &cobra.Command{
	Use:   "check-updates",
	Short: "TODO",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		loader := solver.FilesystemPackageLoader{
			Root:    pkgRoot,
			Context: options.GetVariables(),
		}

		packages, err := solver.NewPackages(&loader)
		if err != nil {
			log.Fatal(err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		l := log.New(log.Writer(), "[check-updates] ", log.Flags())
		if !debug {
			l.SetOutput(ioutil.Discard)
		}

		for _, node := range packages.ToSet() {
			for _, step := range node.Pkg.Steps {
				for _, src := range step.Sources {
					v, url, err := update.Latest(src.URL)
					if err != nil {
						l.Print(err)
						continue
					}
					prefix := "no update"
					if url.String() != src.URL {
						prefix = "update available"
						fmt.Fprintf(w, "%s\t%s\t%s\n", node.Pkg.Name, v, url)
					}
					l.Printf("%s %s %s %s", prefix, node.Pkg.Name, v, url)
				}
			}
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(checkUpdatesCmd)
}
