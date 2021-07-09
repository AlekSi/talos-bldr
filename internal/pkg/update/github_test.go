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

	c := newGitHub(getGitHubToken())

	for source, expected := range map[string]*UpdateInfo{
		// https://github.com/pullmoll/musl-fts/releases has only tags.
		"https://github.com/pullmoll/musl-fts/archive/refs/tags/1.2.6.tar.gz": {
			HasUpdate: true,
			URL:       "https:/github.com/pullmoll/musl-fts/releases/",
		},
		"https://github.com/pullmoll/musl-fts/archive/refs/tags/1.2.7.tar.gz": {
			HasUpdate: false,
			URL:       "https:/github.com/pullmoll/musl-fts/releases/",
		},

		// https://github.com/golang/protobuf/releases has releases without extra assets.
		"https://github.com/golang/protobuf/archive/refs/tags/v1.5.1.tar.gz": {
			HasUpdate: true,
			URL:       "https://github.com/golang/protobuf/releases/",
		},
		"https://github.com/golang/protobuf/archive/refs/tags/v1.5.2.tar.gz": {
			HasUpdate: true,
			URL:       "https://github.com/golang/protobuf/releases/",
		},

		// https://github.com/protocolbuffers/protobuf/releases has releases with extra assets.
		"https://github.com/protocolbuffers/protobuf/releases/download/v3.15.6/protobuf-cpp-3.15.6.tar.gz": {
			HasUpdate: true,
			URL:       "https://github.com/protocolbuffers/protobuf/releases/",
		},
		"https://github.com/protocolbuffers/protobuf/releases/download/v3.17.3/protobuf-cpp-3.17.3.tar.gz": {
			HasUpdate: false,
			URL:       "https://github.com/protocolbuffers/protobuf/releases/",
		},
	} {
		source, expected := source, expected

		u, err := url.Parse(source)
		require.NoError(t, err)
		parts := strings.Split(u.Path, "/")
		owner, repo := parts[1], parts[2]

		t.Run(owner+"/"+repo, func(t *testing.T) {
			t.Parallel()

			// check that source is actually working (with optional redirects)
			resp, err := http.Head(source)
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode)
			require.NoError(t, resp.Body.Close())

			actual, err := c.Latest(context.Background(), source)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}
