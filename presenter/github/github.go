// Package github provides a GitHub API-powered presenter. It supports repositories that are on github.com.
package github

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/google/go-github/github"
	"github.com/shurcooL/Go-Package-Store"
)

// NewPresenter returns a GitHub API-powered presenter.
// httpClient is the HTTP client to be used by the presenter for accessing the GitHub API.
// If httpClient is nil, then http.DefaultClient is used.
func NewPresenter(httpClient *http.Client) gps.Presenter {
	gh := github.NewClient(httpClient)
	gh.UserAgent = "github.com/shurcooL/Go-Package-Store/presenter/github"

	return func(repo *gps.Repo) gps.Presentation {
		switch {
		// Import path begins with "github.com/".
		case strings.HasPrefix(repo.Root, "github.com/"):
			elems := strings.Split(repo.Root, "/")
			if len(elems) != 3 {
				return nil
			}
			return presentGitHubRepo(gh, repo, elems[1], elems[2])
		// gopkg.in package.
		case strings.HasPrefix(repo.Root, "gopkg.in/"):
			githubOwner, githubRepo, err := gopkgInImportPathToGitHub(repo.Root)
			if err != nil {
				return nil
			}
			return presentGitHubRepo(gh, repo, githubOwner, githubRepo)
		// Underlying GitHub remote.
		case strings.HasPrefix(repo.Remote.RepoURL, "https://github.com/"):
			elems := strings.Split(strings.TrimSuffix(repo.Remote.RepoURL[len("https://"):], ".git"), "/")
			if len(elems) != 3 {
				return nil
			}
			return presentGitHubRepo(gh, repo, elems[1], elems[2])
		// Go repo remote has a GitHub mirror repo.
		case strings.HasPrefix(repo.Remote.RepoURL, "https://go.googlesource.com/"):
			repoName := repo.Remote.RepoURL[len("https://go.googlesource.com/"):]
			return presentGitHubRepo(gh, repo, "golang", repoName)
		default:
			return nil
		}
	}
}

type githubPresentation struct {
	repo    *gps.Repo
	ghOwner string
	ghRepo  string

	cc    *github.CommitsComparison
	image template.URL
	err   error
}

func presentGitHubRepo(gh *github.Client, repo *gps.Repo, ghOwner, ghRepo string) gps.Presentation {
	p := &githubPresentation{
		repo:    repo,
		ghOwner: ghOwner,
		ghRepo:  ghRepo,

		image: "https://github.com/images/gravatars/gravatar-user-420.png", // Default fallback.
	}

	// This might take a while.
	if cc, _, err := gh.Repositories.CompareCommits(ghOwner, ghRepo, repo.Local.Revision, repo.Remote.Revision); err == nil {
		p.cc = cc
	} else if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		p.setFirstError(rateLimitError{rateLimitErr})
	} else {
		p.setFirstError(fmt.Errorf("gh.Repositories.CompareCommits: %v", err))
	}

	// Use the repo owner avatar image.
	if user, _, err := gh.Users.Get(ghOwner); err == nil && user.AvatarURL != nil {
		p.image = template.URL(*user.AvatarURL)
	} else if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		p.setFirstError(rateLimitError{rateLimitErr})
	} else {
		p.setFirstError(fmt.Errorf("gh.Users.Get: %v", err))
	}

	return p
}

func (p githubPresentation) Home() *template.URL {
	switch {
	case strings.HasPrefix(p.repo.Root, "github.com/"):
		url := template.URL("https://github.com/" + p.ghOwner + "/" + p.ghRepo)
		return &url
	default:
		url := template.URL("http://" + p.repo.Root)
		return &url
	}
}

func (p githubPresentation) Image() template.URL {
	return p.image
}

func (p githubPresentation) Changes() <-chan gps.Change {
	if p.cc == nil {
		return nil
	}
	out := make(chan gps.Change)
	go func() {
		for i := range p.cc.Commits {
			c := p.cc.Commits[len(p.cc.Commits)-1-i] // Reverse order.
			change := gps.Change{
				Message: firstParagraph(*c.Commit.Message),
				URL:     template.URL(*c.HTMLURL),
			}
			if commentCount := c.Commit.CommentCount; commentCount != nil && *commentCount > 0 {
				change.Comments.Count = *commentCount
				change.Comments.URL = template.URL(*c.HTMLURL + "#comments")
			}
			out <- change
		}
		close(out)
	}()
	return out
}

// firstParagraph returns the first paragraph of text s.
func firstParagraph(s string) string {
	i := strings.Index(s, "\n\n")
	if i == -1 {
		return s
	}
	return s[:i]
}

func (p githubPresentation) Error() error { return p.err }

// setFirstError sets error if it's the first one. It does nothing otherwise.
func (p *githubPresentation) setFirstError(err error) {
	if p.err != nil {
		return
	}
	p.err = err
}

// rateLimitError is an error presentation wrapper for consistent display of *github.RateLimitError.
type rateLimitError struct {
	err *github.RateLimitError
}

func (r rateLimitError) Error() string {
	return fmt.Sprintf("GitHub API rate limit exceeded; it will be reset in %v (but you can set GO_PACKAGE_STORE_GITHUB_TOKEN env var for higher rate limit)", humanize.Time(r.err.Rate.Reset.Time))
}
