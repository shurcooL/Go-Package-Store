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
	cc          *github.CommitsComparison
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

	return p
}

func (this gitHubPresenter) Repo() *pkg.Repo {
	return this.repo
}

func (this gitHubPresenter) HomePage() *template.URL {
	switch {
	case strings.HasPrefix(this.repo.RepoImportPath(), "github.com/"):
		url := template.URL("https://github.com/" + this.gitHubOwner + "/" + this.gitHubRepo)
		return &url
	default:
		url := template.URL("http://" + this.repo.RepoImportPath())
		return &url
	}
}

func (this gitHubPresenter) Image() template.URL {
	// Use the repo owner avatar image.
	if user, _, err := gh.Users.Get(this.gitHubOwner); err == nil && user.AvatarURL != nil {
		return template.URL(*user.AvatarURL)
	}
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}

func (this gitHubPresenter) Changes() <-chan Change {
	if this.cc == nil {
		return nil
	}
	out := make(chan Change)
	go func() {
		for index := range this.cc.Commits {
			change := Change{
				Message: firstParagraph(*this.cc.Commits[len(this.cc.Commits)-1-index].Commit.Message),
				URL:     template.URL(*this.cc.Commits[len(this.cc.Commits)-1-index].HTMLURL),
			}
			if commentCount := this.cc.Commits[len(this.cc.Commits)-1-index].Commit.CommentCount; commentCount != nil && *commentCount > 0 {
				change.Comments.Count = *commentCount
				change.Comments.URL = template.URL(*this.cc.Commits[len(this.cc.Commits)-1-index].HTMLURL + "#comments")
			}
			out <- change
		}
		close(out)
	}()
	return out
}

// firstParagraph returns the first paragraph of a string.
func firstParagraph(s string) string {
	if index := strings.Index(s, "\n\n"); index != -1 {
		return s[:index]
	}
	return s
}

//var gh = github.NewClient(oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ""})))
var gh = github.NewClient(nil)
