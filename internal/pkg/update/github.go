/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package update

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

var (
	getGitHubClientOnce sync.Once
	gitHubClient        *github.Client
)

// getGitHubClient returns configured GitHub client.
func getGitHubClient() *github.Client {
	getGitHubClientOnce.Do(func() {
		token := os.Getenv("BLDR_GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		var httpClient *http.Client
		if token != "" {
			src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
			httpClient = oauth2.NewClient(context.Background(), src)
		}

		gitHubClient = github.NewClient(httpClient)
	})

	return gitHubClient
}

func latestGitHub(ctx context.Context, source string, debugf printfFunc) (*UpdateInfo, error) {
	sourceURL, err := url.Parse(source)
	if err != nil {
		return nil, err
	}

	if sourceURL.Host != "github.com" {
		return nil, fmt.Errorf("unexpected host %q", sourceURL.Host)
	}

	releases, err := getReleases(ctx, sourceURL)
	if err != nil {
		return nil, err
	}

	if len(releases) != 0 {
		return latestGitHubRelease(ctx, releases, sourceURL, debugf)
	}

	tags, err := getTags(ctx, sourceURL)
	if err != nil {
		return nil, err
	}

	return latestGitHubTag(ctx, tags, sourceURL, debugf)
}

func latestGitHubRelease(ctx context.Context, releases []*github.RepositoryRelease, sourceURL *url.URL, debugf printfFunc) (*UpdateInfo, error) {
	// find newest release
	newest := releases[0]
	for _, release := range releases {
		if newest.CreatedAt.Before(release.CreatedAt.Time) {
			newest = release
		}
	}

	debugf("Newest release: %+v", newest)

	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]
	res := &UpdateInfo{
		URL: fmt.Sprintf("https://github.com/%s/%s/releases/", owner, repo),
	}

	// update is available if the newest release doesn't have source in their assets download URLs
	source := sourceURL.String()
	for _, asset := range newest.Assets {
		if asset.GetBrowserDownloadURL() == source {
			res.HasUpdate = false
			return res, nil
		}
	}

	res.HasUpdate = true
	return res, nil
}

func latestGitHubTag(ctx context.Context, tags []*github.RepositoryTag, sourceURL *url.URL, debugf printfFunc) (*UpdateInfo, error) {
	return &UpdateInfo{}, nil

	// var latestVersion semver.Version
	// var latestURL string
	// for _, tag := range tags {
	// 	v, err := semver.NewVersion(*tag.Name)
	// 	if err != nil {
	// 		debugf("%s", err)
	// 		continue
	// 	}

	// 	if v.Prerelease() != "" {
	// 		debugf("%s - skipping pre-release", v)
	// 		continue
	// 	}

	// 	if latestVersion.LessThan(v) {
	// 		latestVersion = *v

	// 		// latestTag.TarballURL / ZipballURL are not good enough, construct URL manually
	// 		latestURL = fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", owner, repo, *tag.Name)
	// 	}
	// }

	// var currentVersion string
	// if current, _ := extractVersion(source); current != nil {
	// 	currentVersion = current.String()
	// }

	// _ = currentVersion
	// _ = latestURL

	// res := &UpdateInfo{
	// 	HasUpdate: true,
	// 	// CurrentVersion: currentVersion,
	// 	// LatestVersion:  latestVersion.String(),
	// 	// LatestURL:      latestURL,
	// }
	// return res, nil
}

func getReleases(ctx context.Context, sourceURL *url.URL) ([]*github.RepositoryRelease, error) {
	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]

	opts := &github.ListOptions{
		PerPage: 100,
	}

	res, _, err := getGitHubClient().Repositories.ListReleases(ctx, owner, repo, opts)
	if len(res) == 100 {
		return res, fmt.Errorf("got %d results, pagination should be implemented", len(res))
	}
	return res, err
}

func getTags(ctx context.Context, sourceURL *url.URL) ([]*github.RepositoryTag, error) {
	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]

	opts := &github.ListOptions{
		PerPage: 100,
	}

	res, _, err := getGitHubClient().Repositories.ListTags(ctx, owner, repo, opts)
	if len(res) == 100 {
		return res, fmt.Errorf("got %d results, pagination should be implemented", len(res))
	}
	return res, err
}
