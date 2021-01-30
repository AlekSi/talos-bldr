/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractVersion(t *testing.T) {
	for s, expected := range map[string]string{
		"https://ftp.gnu.org/gnu/automake/automake-1.16.tar.gz":   "1.16.0",
		"https://ftp.gnu.org/gnu/automake/automake-1.16.1.tar.xz": "1.16.1",
	} {
		t.Run(s, func(t *testing.T) {
			actualV := extractVersion(s)
			assert.Equal(t, expected, actualV.String())
		})
	}
}
