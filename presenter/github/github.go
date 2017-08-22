// Package github provides a GitHub API-powered presenter. It supports repositories that are on github.com.
package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/google/go-github/github"
	"github.com/shurcooL/Go-Package-Store/presenter"
)

// NewPresenter returns a GitHub API-powered presenter.
// httpClient is the HTTP client to be used by the presenter for accessing the GitHub API.
// If httpClient is nil, then http.DefaultClient is used.
func NewPresenter(httpClient *http.Client) presenter.Presenter {
	gh := github.NewClient(httpClient)
	gh.UserAgent = "github.com/shurcooL/Go-Package-Store/presenter/github"

	return func(ctx context.Context, repo presenter.Repo) *presenter.Presentation {
		switch {
		// Import path begins with "github.com/".
		case strings.HasPrefix(repo.Root, "github.com/"):
			elems := strings.Split(repo.Root, "/")
			if len(elems) != 3 {
				return nil
			}
			return presentGitHubRepo(ctx, gh, repo, elems[1], elems[2])
		// gopkg.in package.
		case strings.HasPrefix(repo.Root, "gopkg.in/"):
			githubOwner, githubRepo, err := gopkgInImportPathToGitHub(repo.Root)
			if err != nil {
				return nil
			}
			return presentGitHubRepo(ctx, gh, repo, githubOwner, githubRepo)
		// Underlying GitHub remote.
		case strings.HasPrefix(repo.RepoURL, "https://github.com/"):
			elems := strings.Split(strings.TrimSuffix(repo.RepoURL[len("https://"):], ".git"), "/")
			if len(elems) != 3 {
				return nil
			}
			return presentGitHubRepo(ctx, gh, repo, elems[1], elems[2])
		// Go repo remote has a GitHub mirror repo.
		case strings.HasPrefix(repo.RepoURL, "https://go.googlesource.com/"):
			repoName := repo.RepoURL[len("https://go.googlesource.com/"):]
			return presentGitHubRepo(ctx, gh, repo, "golang", repoName)
		// upspin.io.
		case strings.HasPrefix(repo.RepoURL, "https://upspin.googlesource.com/"):
			repoName := repo.RepoURL[len("https://upspin.googlesource.com/"):]
			return presentGitHubRepo(ctx, gh, repo, "upspin", repoName)
		default:
			return nil
		}
	}
}

func presentGitHubRepo(ctx context.Context, gh *github.Client, repo presenter.Repo, ghOwner, ghRepo string) *presenter.Presentation {
	p := &presenter.Presentation{
		HomeURL:  "https://" + repo.Root,
		ImageURL: "https://github.com/images/gravatars/gravatar-user-420.png", // Default fallback.
	}

	// This might take a while.
	if cc, _, err := gh.Repositories.CompareCommits(ctx, ghOwner, ghRepo, repo.LocalRevision, repo.RemoteRevision); err == nil {
		p.Changes = extractChanges(cc)
	} else if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		setFirstError(p, rateLimitError{rateLimitErr})
	} else {
		setFirstError(p, fmt.Errorf("gh.Repositories.CompareCommits: %v", err))
	}

	// Use the repo owner avatar image.
	if repo, _, err := gh.Repositories.Get(ctx, ghOwner, ghRepo); err == nil && repo.Owner != nil && repo.Owner.AvatarURL != nil {
		p.ImageURL = *repo.Owner.AvatarURL
	} else if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		setFirstError(p, rateLimitError{rateLimitErr})
	} else {
		setFirstError(p, fmt.Errorf("gh.Repositories.Get: %v", err))
	}

	return p
}

func extractChanges(cc *github.CommitsComparison) []presenter.Change {
	var cs []presenter.Change
	for i := range cc.Commits {
		c := cc.Commits[len(cc.Commits)-1-i] // Reverse order.
		change := presenter.Change{
			Message: firstParagraph(*c.Commit.Message),
			URL:     *c.HTMLURL,
		}
		if commentCount := c.Commit.CommentCount; commentCount != nil && *commentCount > 0 {
			change.Comments.Count = *commentCount
			change.Comments.URL = *c.HTMLURL + "#comments"
		}
		cs = append(cs, change)
	}
	return cs
}

// firstParagraph returns the first paragraph of text s.
func firstParagraph(s string) string {
	i := strings.Index(s, "\n\n")
	if i == -1 {
		return s
	}
	return s[:i]
}

// rateLimitError is an error presentation wrapper for consistent display of *github.RateLimitError.
type rateLimitError struct {
	err *github.RateLimitError
}

func (r rateLimitError) Error() string {
	return fmt.Sprintf("GitHub API rate limit exceeded; it will be reset in %v (but you can set GO_PACKAGE_STORE_GITHUB_TOKEN env var for higher rate limit)", humanize.Time(r.err.Rate.Reset.Time))
}

// setFirstError sets error if it's the first one. It does nothing otherwise.
func setFirstError(p *presenter.Presentation, err error) {
	if p.Error != nil {
		return
	}
	p.Error = err
}
