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
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

var (
	getClientOnce   sync.Once
	getClientClient *github.Client
)

func getClient() *github.Client {
	getClientOnce.Do(func() {
		token := os.Getenv("BLDR_GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		var httpClient *http.Client
		if token != "" {
			src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
			httpClient = oauth2.NewClient(context.Background(), src)
		}

		getClientClient = github.NewClient(httpClient)
	})

	return getClientClient
}

func latestGithub(ctx context.Context, source string, debugf printfFunc) (*UpdateInfo, error) {
	sourceURL, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]

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

func getAllTags(ctx context.Context, owner, repo string) ([]*github.RepositoryTag, error) {
	var res []*github.RepositoryTag
	opts := &github.ListOptions{
		PerPage: 100,
	}
	for {
		tags, resp, err := getClient().Repositories.ListTags(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		res = append(res, tags...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return res, nil
}
