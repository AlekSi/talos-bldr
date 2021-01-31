/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
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
		".diff": {},
		".gz":   {},
		".src":  {},
		".tar":  {},
		".xdp":  {},
		".xz":   {},
	}
)

func Latest(source string) (string, *url.URL, error) {
	u, err := url.Parse(source)
	if err != nil {
		return "", nil, err
	}
	dirU := *u
	dirU.Path = filepath.Dir(u.Path)

	resp, err := http.Get(dirU.String())
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	versions := parseHTML(u, resp.Body)
	if len(versions) == 0 {
		return "", nil, fmt.Errorf("no versions found in HTML")
	}
	latest := semver.MustParse("0.0.0")
	var latestURL *url.URL
	for u, v := range versions {
		if latest.LessThan(v) {
			latestURL = u
			latest = v
		}
	}

	return latest.String(), latestURL, nil
}

func extractVersion(s string) *semver.Version {
	// extract file name
	if u, _ := url.Parse(s); u != nil {
		s = u.Path
	}
	s = filepath.Base(s)

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
		return nil
	}
	s = s[i:]

	res, _ := semver.NewVersion(s)
	return res
}

func parseHTML(sourceURL *url.URL, html io.Reader) map[*url.URL]*semver.Version {
	d := xml.NewDecoder(html)
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
			if v := extractVersion(attr.Value); v != nil {
				if v.Prerelease() != "" {
					continue
				}

				u, err := url.Parse(attr.Value)
				if err != nil {
					continue
				}
				if u.Host == "" {
					u = sourceURL.ResolveReference(u)
				}

				res[u] = v
			}
		}
	}

	return res
}
