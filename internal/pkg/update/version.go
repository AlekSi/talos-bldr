/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"strings"

	"github.com/Masterminds/semver"
)

var (
	extensions = []string{".tar", ".gz", ".xz"}
)

func extractVersion(s string) *semver.Version {
	// extract file name
	s = strings.TrimPrefix(s, "https://ftp.gnu.org/gnu/automake/automake-")

	// remove extensions
	found := true
	for found {
		found = false
		for _, ext := range extensions {
			if strings.HasSuffix(s, ext) {
				s = strings.TrimSuffix(s, ext)
				found = true
			}
		}
	}

	res, _ := semver.NewVersion(s)
	return res
}
