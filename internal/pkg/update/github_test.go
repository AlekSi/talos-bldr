/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// https://github.com/golang/protobuf/archive/v1.4.2.tar.gz
// https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protobuf-cpp-3.13.0.tar.gz

func parseURL(t *testing.T, s string) *url.URL {
	u, err := url.Parse(s)
	require.NoError(t, err)
	return u
}

func TestLatestGithub(t *testing.T) {
	// TODO Ask / decide how we call those tests, should they be run by default, and how.

	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	for source, expected := range map[string]*UpdateInfo{
		// no real releases at https://github.com/pullmoll/musl-fts/releases
		"https://github.com/pullmoll/musl-fts/archive/v1.2.6/musl-fts-1.2.6.tar.gz": {
			CurrentVersion: "1.2.6",
			LatestVersion:  "1.2.7",
			LatestURL:      parseURL(t, "https://github.com/pullmoll/musl-fts/archive/v1.2.6.tar.gz"),
		},
		"https://github.com/pullmoll/musl-fts/archive/v1.2.6.tar.gz": {
			CurrentVersion: "1.2.6",
			LatestVersion:  "1.2.7",
			LatestURL:      parseURL(t, "https://github.com/pullmoll/musl-fts/archive/v1.2.6.tar.gz"),
		},

		// "https://github.com/golang/protobuf/archive/v1.4.2.tar.gz": {
		// 	CurrentVersion: "1.4.2",
		// 	LatestVersion:  "1.4.3",
		// 	LatestURL:      parseURL(t, "https://github.com/golang/protobuf/archive/v1.4.3.tar.gz"),
		// },

		// "https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protobuf-cpp-3.13.0.tar.gz": {
		// 	CurrentVersion: "3.13.0",
		// 	LatestVersion:  "3.14.0",
		// 	LatestURL:      parseURL(t, "https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protobuf-cpp-3.14.0.tar.gz"),
		// },
	} {
		u, err := url.Parse(source)
		require.NoError(t, err)

		parts := strings.Split(u.Path, "/")
		owner, repo := parts[1], parts[2]
		name := owner + "/" + repo
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := latestGithub(context.Background(), source, t.Logf)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}
