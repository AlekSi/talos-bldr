/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// https://github.com/golang/protobuf/archive/v1.4.2.tar.gz
// https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protobuf-cpp-3.13.0.tar.gz

func TestLatestGithub(t *testing.T) {
	// TODO Ask / decide how we call those tests, should they be run by default, and how.

	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	for source, expected := range map[string]*UpdateInfo{
		// https://github.com/pullmoll/musl-fts/releases has only tags;
		// both source forms should be handled.
		"https://github.com/pullmoll/musl-fts/archive/v1.2.6.tar.gz": {
			CurrentVersion: "1.2.6",
			LatestVersion:  "1.2.7",
			LatestURL:      "https://github.com/pullmoll/musl-fts/archive/v1.2.7.tar.gz",
		},
		"https://github.com/pullmoll/musl-fts/archive/v1.2.6/DUMMY.tar.gz": {
			CurrentVersion: "",
			LatestVersion:  "1.2.7",
			LatestURL:      "https://github.com/pullmoll/musl-fts/archive/v1.2.7.tar.gz",
		},

		// https://github.com/golang/protobuf/releases has releases without extra assets;
		// both source froms should be handled.
		"https://github.com/golang/protobuf/archive/v1.4.2.tar.gz": {
			CurrentVersion: "1.4.2",
			LatestVersion:  "1.4.3",
			LatestURL:      "https://github.com/golang/protobuf/archive/v1.4.3.tar.gz",
		},
		"https://github.com/golang/protobuf/archive/v1.4.2/DUMMY.tar.gz": {
			CurrentVersion: "",
			LatestVersion:  "1.4.3",
			LatestURL:      "https://github.com/golang/protobuf/archive/v1.4.3.tar.gz",
		},

		// https://github.com/protocolbuffers/protobuf/releases has releases with extra assets.
		"https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protobuf-cpp-3.13.0.tar.gz": {
			CurrentVersion: "3.13.0",
			LatestVersion:  "3.15.6",
			LatestURL:      "https://github.com/protocolbuffers/protobuf/releases/download/v3.15.6/protobuf-cpp-3.15.6.tar.gz",
		},
	} {
		source, expected := source, expected

		// check that source is actually working
		resp, err := http.Head(source)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.NoError(t, resp.Body.Close())

		u, err := url.Parse(source)
		require.NoError(t, err)
		parts := strings.Split(u.Path, "/")
		owner, repo := parts[1], parts[2]

		t.Run(owner+"/"+repo, func(t *testing.T) {
			t.Parallel()

			actual, err := latestGithub(context.Background(), source, t.Logf)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}
