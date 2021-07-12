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
	"time"

	"github.com/google/go-github/v35/github"
	"golang.org/x/oauth2"
)

func gitHubTokenFromEnv() string {
	token := os.Getenv("BLDR_GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	return token
}

type gitHub struct {
	c *github.Client
}

func newGitHub(token string) *gitHub {
	var c *http.Client
	if token != "" {
		src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		c = oauth2.NewClient(context.Background(), src)
	}

	return &gitHub{
		c: github.NewClient(c),
	}
}

// Latest returns information about available update.
func (g *gitHub) Latest(ctx context.Context, source string) (*LatestInfo, error) {
	sourceURL, err := url.Parse(source)
	if err != nil {
		return nil, err
	}

	if sourceURL.Host != "github.com" {
		panic(fmt.Sprintf("unexpected host %q", sourceURL.Host))
	}

	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]

	v, err := extractVersion(source)
	if err != nil {
		return nil, err
	}

	considerPrereleases := v.Prerelease() != ""

	releases, err := g.getReleases(ctx, owner, repo)
	if err != nil {
		return nil, g.wrapGitHubError(err)
	}

	if len(releases) != 0 {
		return g.findLatestRelease(ctx, releases, sourceURL, considerPrereleases)
	}

	tags, err := g.getTags(ctx, owner, repo)
	if err != nil {
		return nil, g.wrapGitHubError(err)
	}

	return g.findLatestTag(ctx, tags, sourceURL, considerPrereleases)
}

// findLatestRelease returns information about latest released version.
func (g *gitHub) findLatestRelease(ctx context.Context, releases []*github.RepositoryRelease, sourceURL *url.URL, considerPrereleases bool) (*LatestInfo, error) {
	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]

	// find newest release
	var newest *github.RepositoryRelease
	for _, release := range releases {
		if release.GetPrerelease() && !considerPrereleases {
			continue
		}

		if newest == nil || newest.CreatedAt.Before(release.CreatedAt.Time) {
			newest = release
		}
	}

	if newest == nil {
		return nil, fmt.Errorf("no release found")
	}

	res := &LatestInfo{
		BaseURL: fmt.Sprintf("https://github.com/%s/%s/releases/", owner, repo),
	}

	source := sourceURL.String()

	// treat releases without extra assets as tags
	if len(newest.Assets) == 0 {
		// update is available if the release doesn't have the same tarball URL
		res.LatestURL = g.getTagGZ(owner, repo, newest.GetTagName())
		res.HasUpdate = res.LatestURL != source
		return res, nil
	}

	// update is available if the newest release doesn't have source in their assets download URLs
	for _, asset := range newest.Assets {
		if asset.GetBrowserDownloadURL() == source {
			res.HasUpdate = false
			res.LatestURL = source
			return res, nil
		}
	}

	// we don't know correct asset URL
	res.HasUpdate = true
	return res, nil
}

// findLatestTag returns information about latest tagged version.
func (g *gitHub) findLatestTag(ctx context.Context, tags []*github.RepositoryTag, sourceURL *url.URL, considerPrereleases bool) (*LatestInfo, error) {
	parts := strings.Split(sourceURL.Path, "/")
	owner, repo := parts[1], parts[2]

	// find newest tag
	var newest *github.RepositoryTag
	var newestDate time.Time
	for _, tag := range tags {
		v, err := extractVersion(tag.GetName())
		if err != nil {
			return nil, err
		}

		if v.Prerelease() != "" && !considerPrereleases {
			continue
		}

		tagDate, err := g.getCommitTime(ctx, owner, repo, tag.GetCommit().GetSHA())
		if err != nil {
			return nil, err
		}

		if newest == nil || newestDate.Before(tagDate) {
			newest = tag
			newestDate = tagDate
		}
	}

	if newest == nil {
		return nil, fmt.Errorf("no tag found")
	}

	res := &LatestInfo{
		BaseURL:   fmt.Sprintf("https://github.com/%s/%s/releases/", owner, repo),
		LatestURL: g.getTagGZ(owner, repo, newest.GetName()),
	}

	// update is available if the newest tag doesn't have the same tarball URL
	res.HasUpdate = res.LatestURL != sourceURL.String()
	return res, nil
}

// getReleases returns all releases.
func (g *gitHub) getReleases(ctx context.Context, owner, repo string) ([]*github.RepositoryRelease, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	res, _, err := g.c.Repositories.ListReleases(ctx, owner, repo, opts)
	if len(res) == 100 {
		return res, fmt.Errorf("got %d results, pagination should be implemented", len(res))
	}
	return res, err
}

// getTags returns all tags.
func (g *gitHub) getTags(ctx context.Context, owner, repo string) ([]*github.RepositoryTag, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	res, _, err := g.c.Repositories.ListTags(ctx, owner, repo, opts)
	if len(res) == 100 {
		return res, fmt.Errorf("got %d results, pagination should be implemented", len(res))
	}
	return res, err
}

// getCommitTime returns commit's time.
func (g *gitHub) getCommitTime(ctx context.Context, owner, repo, sha string) (time.Time, error) {
	commit, _, err := g.c.Repositories.GetCommit(ctx, owner, repo, sha)
	if err != nil {
		return time.Time{}, err
	}

	t := commit.GetCommit().GetCommitter().GetDate()
	if t.IsZero() {
		return time.Time{}, fmt.Errorf("no commit date")
	}

	return t, nil
}

// getTagTarball returns .tar.gz URL.
// API's GetTarballURL is not good enough.
func (g *gitHub) getTagGZ(owner, repo, name string) string {
	return fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", owner, repo, name)
}

func (g *gitHub) wrapGitHubError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(*github.RateLimitError); ok {
		err = fmt.Errorf("%w\nSet `BLDR_GITHUB_TOKEN` or `GITHUB_TOKEN` environment variable.", err)
	}

	return err
}
