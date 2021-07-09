/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/talos-systems/bldr/internal/pkg/solver"
	"github.com/talos-systems/bldr/internal/pkg/update"
)

type packageInfo struct {
	file   string
	source string
}

type updateInfo struct {
	file string
	*update.UpdateInfo
}

var all bool

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

		const concurrency = 10
		var wg sync.WaitGroup
		sources := make(chan *packageInfo)
		updates := make(chan *updateInfo)
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for src := range sources {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					res, err := update.Latest(ctx, src.source, l.Printf)
					cancel()
					if err != nil {
						l.Print(err)
						continue
					}
					updates <- &updateInfo{
						file:       src.file,
						UpdateInfo: res,
					}
				}
			}()
		}

		var res []updateInfo
		done := make(chan struct{})
		go func() {
			for update := range updates {
				res = append(res, *update)
			}
			close(done)
		}()

		for _, node := range packages.ToSet() {
			for _, step := range node.Pkg.Steps {
				for _, src := range step.Sources {
					sources <- &packageInfo{
						file:   node.Pkg.FileName,
						source: src.URL,
					}
				}
			}
		}
		close(sources)
		wg.Wait()
		close(updates)
		<-done

		sort.Slice(res, func(i, j int) bool { return res[i].file < res[j].file })

		for _, u := range res {
			if all || u.HasUpdate {
				fmt.Fprintf(w, "%s\t%t\t%s\n", u.file, u.HasUpdate, u.BaseURL)
			}
		}

		// v, url, err := update.Latest(src.URL)
		// if err != nil {
		// 	l.Print(err)
		// 	continue
		// }
		// prefix := "no update"
		// if url.String() != src.URL {
		// 	prefix = "update available"
		// 	fmt.Fprintf(w, "%s\t%s\t%s\n", node.Pkg.Name, v, url)
		// }
		// l.Printf("%s %s %s %s", prefix, node.Pkg.Name, v, url)

		w.Flush()
	},
}

func init() {
	checkUpdatesCmd.Flags().BoolVarP(&all, "all", "a", false, "")
	rootCmd.AddCommand(checkUpdatesCmd)
}
