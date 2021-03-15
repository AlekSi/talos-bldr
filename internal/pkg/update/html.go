/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

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
