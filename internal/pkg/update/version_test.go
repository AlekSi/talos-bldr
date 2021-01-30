/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractVersion(t *testing.T) {
	for s, expected := range map[string]string{
		"https://dl.google.com/go/go1.15.7.src.tar.gz":            "1.15.7",
		"https://ftp.gnu.org/gnu/automake/automake-1.16.1.tar.xz": "1.16.1",
		"https://ftp.gnu.org/gnu/automake/automake-1.16.tar.gz":   "1.16.0",
	} {
		t.Run(s, func(t *testing.T) {
			actualV := extractVersion(s)
			require.NotNil(t, actualV)
			assert.Equal(t, expected, actualV.String())
		})
	}
}

func TestParseHTML(t *testing.T) {
	matches, err := filepath.Glob("testdata/*.json")
	require.NoError(t, err)

	for _, match := range matches {
		name := strings.TrimPrefix(strings.TrimSuffix(match, ".json"), "testdata/")
		t.Run(name, func(t *testing.T) {
			// read expected versions
			b, err := ioutil.ReadFile(match)
			require.NoError(t, err)
			var expected []string
			err = json.Unmarshal(b, &expected)
			require.NoError(t, err)

			// read actual versions
			r, err := os.Open(strings.TrimSuffix(match, ".json") + ".html")
			require.NoError(t, err)
			defer r.Close()
			actualV := parseHTML(r)
			require.NotNil(t, actualV)
			actual := make([]string, len(actualV))
			for i, a := range actualV {
				actual[i] = a.String()
			}

			assert.Equal(t, expected, actual)
		})
	}
}
