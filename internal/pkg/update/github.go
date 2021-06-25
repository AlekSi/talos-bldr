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

	"github.com/Masterminds/semver"
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

func latestGithub(ctx context.Context, source string, debugf printfFunc) (*UpdateInfo, error) {
	sourceURL, err := url.Parse(source)
	if err != nil {
		return nil, err
	}

	if sourceURL.Host != "github.com" {
		return nil, fmt.Errorf("unexpected host %q", sourceURL.Host)
	}

	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]

	releases, err := getAllReleases(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	// TODO
	_ = releases

	tags, err := getAllTags(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	var latestVersion semver.Version
	var latestURL string
	for _, tag := range tags {
		v, err := semver.NewVersion(*tag.Name)
		if err != nil {
			debugf("%s", err)
			continue
		}

		if v.Prerelease() != "" {
			debugf("%s - skipping pre-release", v)
			continue
		}

		if latestVersion.LessThan(v) {
			latestVersion = *v

			// latestTag.TarballURL / ZipballURL are not good enough, construct URL manually
			latestURL = fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", owner, repo, *tag.Name)
		}
	}

	var currentVersion string
	if current, _ := extractVersion(source); current != nil {
		currentVersion = current.String()
	}
	res := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion.String(),
		LatestURL:      latestURL,
	}
	return res, nil
}

func getAllReleases(ctx context.Context, owner, repo string) ([]*github.RepositoryRelease, error) {
	var res []*github.RepositoryRelease
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		page, resp, err := getGitHubClient().Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		res = append(res, page...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return res, nil
}

func getAllReleaseAssets(ctx context.Context, owner, repo string, releaseID int64) ([]*github.ReleaseAsset, error) {
	var res []*github.ReleaseAsset
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		page, resp, err := getGitHubClient().Repositories.ListReleaseAssets(ctx, owner, repo, releaseID, opts)
		if err != nil {
			return nil, err
		}
		res = append(res, page...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return res, nil
}

func getAllTags(ctx context.Context, owner, repo string) ([]*github.RepositoryTag, error) {
	var res []*github.RepositoryTag
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		page, resp, err := getGitHubClient().Repositories.ListTags(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		res = append(res, page...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return res, nil
}
