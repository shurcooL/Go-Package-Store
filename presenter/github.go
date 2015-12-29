package presenter

import (
	"html/template"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/shurcooL/Go-Package-Store/pkg"
)

type gitHubPresenter struct {
	repo        *pkg.Repo
	gitHubOwner string
	gitHubRepo  string

	cc    *github.CommitsComparison
	image template.URL
}

func newGitHubPresenter(repo *pkg.Repo, gitHubOwner, gitHubRepo string) *gitHubPresenter {
	p := &gitHubPresenter{
		repo:        repo,
		gitHubOwner: gitHubOwner,
		gitHubRepo:  gitHubRepo,
	}

	// This might take a while.
	if cc, _, err := gh.Repositories.CompareCommits(gitHubOwner, gitHubRepo, repo.Local.Revision, repo.Remote.Revision); err == nil {
		p.cc = cc
	} else {
		log.Println("warning: gh.Repositories.CompareCommits:", err)
	}

	// Use the repo owner avatar image.
	if user, _, err := gh.Users.Get(gitHubOwner); err == nil && user.AvatarURL != nil {
		p.image = template.URL(*user.AvatarURL)
	} else {
		p.image = "https://github.com/images/gravatars/gravatar-user-420.png"
	}

	return p
}

func (p gitHubPresenter) Repo() *pkg.Repo {
	return p.repo
}

func (p gitHubPresenter) HomePage() *template.URL {
	switch {
	case strings.HasPrefix(p.repo.Root, "github.com/"):
		url := template.URL("https://github.com/" + p.gitHubOwner + "/" + p.gitHubRepo)
		return &url
	default:
		url := template.URL("http://" + p.repo.Root)
		return &url
	}
}

func (p gitHubPresenter) Image() template.URL {
	return p.image
}

func (p gitHubPresenter) Changes() <-chan Change {
	if p.cc == nil {
		return nil
	}
	out := make(chan Change)
	go func() {
		for index := range p.cc.Commits {
			change := Change{
				Message: firstParagraph(*p.cc.Commits[len(p.cc.Commits)-1-index].Commit.Message),
				URL:     template.URL(*p.cc.Commits[len(p.cc.Commits)-1-index].HTMLURL),
			}
			if commentCount := p.cc.Commits[len(p.cc.Commits)-1-index].Commit.CommentCount; commentCount != nil && *commentCount > 0 {
				change.Comments.Count = *commentCount
				change.Comments.URL = template.URL(*p.cc.Commits[len(p.cc.Commits)-1-index].HTMLURL + "#comments")
			}
			out <- change
		}
		close(out)
	}()
	return out
}

// firstParagraph returns the first paragraph of a string.
func firstParagraph(s string) string {
	index := strings.Index(s, "\n\n")
	if index == -1 {
		return s
	}
	return s[:index]
}

//var gh = github.NewClient(oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ""})))
var gh = github.NewClient(nil)
