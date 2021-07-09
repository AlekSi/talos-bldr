/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/v35/github"
)

type UpdateInfo struct {
	// HasUpdate is true if there seems to be an update available.
	HasUpdate bool
	// BaseURL may contain base URL for releases.
	BaseURL string
	// LatestURL may contain URL for the latest asset.
	LatestURL string
}

type printfFunc func(format string, v ...interface{})

func Latest(ctx context.Context, source string, debugf printfFunc) (*UpdateInfo, error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}

	switch u.Host {
	case "github.com":
		res, err := newGitHub(getGitHubToken()).Latest(ctx, source)
		if _, ok := err.(*github.RateLimitError); ok {
			err = fmt.Errorf("%w\nSet `BLDR_GITHUB_TOKEN` or `GITHUB_TOKEN` environment variable.", err)
		}
		return res, err

	default:
		return nil, fmt.Errorf("unhandled host %q", u.Host)
	}
}
