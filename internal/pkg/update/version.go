/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"encoding/xml"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
)

var (
	extensions = map[string]bool{
		".diff": true,
		".gz":   true,
		".src":  true,
		".tar":  true,
		".xdp":  true,
		".xz":   true,
	}
)

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
		if found = extensions[ext]; found {
			s = strings.TrimSuffix(s, ext)
		}
	}

	// remove name prefix
	i := strings.IndexAny(s, "0123456789")
	if i < 0 {
		return nil
	}
	s = s[i:]

	res, _ := semver.NewVersion(s)
	return res
}

func parseHTML(r io.Reader) []*semver.Version {
	d := xml.NewDecoder(r)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity

	var res []*semver.Version
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
				res = append(res, v)
			}
		}
	}

	return res
}
