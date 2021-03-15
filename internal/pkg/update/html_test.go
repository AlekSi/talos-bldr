/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

func TestParseHTML(t *testing.T) {
	type testdata struct {
		SourceURL string            `json:"source_url"`
		Versions  map[string]string `json:"versions"`
	}

	matches, err := filepath.Glob("testdata/*.json")
	require.NoError(t, err)

	for _, match := range matches {
		name := strings.TrimPrefix(strings.TrimSuffix(match, ".json"), "testdata/")
		t.Run(name, func(t *testing.T) {
			// read testdata
			b, err := ioutil.ReadFile(match)
			require.NoError(t, err)
			var td testdata
			err = json.Unmarshal(b, &td)
			require.NoError(t, err)
			sourceURL, err := url.Parse(td.SourceURL)
			require.NoError(t, err)

			// parse HTML
			r, err := os.Open(strings.TrimSuffix(match, ".json") + ".html")
			require.NoError(t, err)
			defer r.Close()
			actual := parseHTML(sourceURL, r)
			require.NotNil(t, actual)
			versions := make(map[string]string, len(actual))
			for u, v := range actual {
				versions[u.String()] = v.String()
			}

			assert.Equal(t, td.Versions, versions)
		})
	}
}
