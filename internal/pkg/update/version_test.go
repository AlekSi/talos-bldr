/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractVersion(t *testing.T) {
	for s, expected := range map[string]string{
		"https://dl.google.com/go/go1.15.7.src.tar.gz":    "1.15.7",
		"https://ftp.pcre.org/pub/pcre/pcre-8.43.tar.bz2": "8.43.0",

		"automake-1.16.tar.gz":                                    "1.16.0",
		"automake-1.16.1.tar.xz":                                  "1.16.1",
		"https://ftp.gnu.org/gnu/automake/automake-1.16.tar.gz":   "1.16.0",
		"https://ftp.gnu.org/gnu/automake/automake-1.16.1.tar.xz": "1.16.1",
	} {
		t.Run(s, func(t *testing.T) {
			actualV, err := extractVersion(s)
			require.NoError(t, err)
			require.NotNil(t, actualV)
			assert.Equal(t, expected, actualV.String())
		})
	}
}
