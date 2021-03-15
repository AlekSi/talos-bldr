/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"context"
	"fmt"
	"net/url"
)

type UpdateInfo struct {
	// Current version, as extracted from the source URL.
	CurrentVersion string
	// Latest version, as determined by the updater.
	LatestVersion string
	// Latest version's full absolute URL of the source.
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
		return latestGithub(ctx, source, debugf)
	default:
		return nil, fmt.Errorf("unhandled host %q", u.Host)
	}
}
