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

	"github.com/AlekSi/pointer"
	"github.com/Masterminds/semver"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

type printfFunc func(format string, v ...interface{})

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
	// sanity check
	sourceURL, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	if sourceURL.Host != "github.com" {
		return nil, fmt.Errorf("unexpected host %q", sourceURL.Host)
	}

	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]
	tags, _, err := getClient().Repositories.ListTags(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}

	var latestVersion *semver.Version
	var latestTag *github.RepositoryTag
	var latestURL *url.URL
	for _, latestTag = range tags {
		latestVersion, err = semver.NewVersion(*latestTag.Name)
		if err != nil {
			debugf("%s", err)
			continue
		}

		if latestVersion.Prerelease() != "" {
			debugf("%s - skipping pre-release", latestVersion)
			continue
		}

		latestURL, err = url.Parse(pointer.GetString(latestTag.TarballURL))
		if err != nil {
			debugf("%s", err)
			continue
		}

		break
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
