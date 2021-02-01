/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
)

var (
	extensions = map[string]struct{}{
		".bz2":  {},
		".diff": {},
		".gz":   {},
		".orig": {},
		".src":  {},
		".tar":  {},
		".xdp":  {},
		".xz":   {},
	}
)

type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	LatestURL      *url.URL
}

func Latest(ctx context.Context, source string) (*UpdateInfo, error) {
	currentVersion, err := extractVersion(source)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	dirU := *u
	dirU.Path = filepath.Dir(u.Path)

	req, err := http.NewRequestWithContext(ctx, "GET", dirU.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	versions := parseHTML(u, resp.Body)
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found in HTML")
	}
	latest := semver.MustParse("0.0.0")
	var latestURL *url.URL
	for u, v := range versions {
		if latest.LessThan(v) {
			latestURL = u
			latest = v
		}
	}

	return &UpdateInfo{
		CurrentVersion: currentVersion.String(),
		LatestVersion:  latest.String(),
		LatestURL:      latestURL,
	}, nil
}

// extractVersion extracts SemVer version from file name or URL.
// If version can't be extracted, nil is returned.
func extractVersion(s string) (*semver.Version, error) {
	// extract file name
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	s = filepath.Base(u.Path)

	// remove common extensions
	found := true
	for found {
		ext := filepath.Ext(s)
		if _, found = extensions[ext]; found {
			s = strings.TrimSuffix(s, ext)
		}
	}

	// remove package name, keep only version
	i := strings.IndexAny(s, "0123456789")
	if i < 0 {
		return nil, fmt.Errorf("failed to remove package name from %q", s)
	}
	s = s[i:]

	res, err := semver.NewVersion(s)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", s, err)
	}
	return res, nil
}

func parseHTML(pageURL *url.URL, pageHTML io.Reader) map[*url.URL]*semver.Version {
	d := xml.NewDecoder(pageHTML)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity

	res := make(map[*url.URL]*semver.Version)
	for {
		t, err := d.Token()
		if err != nil {
			break
		}

		el, ok := t.(xml.StartElement)
		if !ok {
			continue
		}

		if el.Name.Local != "a" {
			continue
		}
		for _, attr := range el.Attr {
			if attr.Name.Local != "href" {
				continue
			}

			v, err := extractVersion(attr.Value)
			if err != nil {
				continue
			}

			if v.Prerelease() != "" {
				continue
			}

			u, err := url.Parse(attr.Value)
			if err != nil {
				continue
			}
			if u.Host == "" {
				u = pageURL.ResolveReference(u)
			}

			res[u] = v
		}
	}

	return res
}
