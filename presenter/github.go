package presenter

import (
	"html/template"
	"strings"

	"github.com/google/go-github/github"
	"github.com/shurcooL/go/exp/13"
	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/shurcooL/go/gists/gist7802150"
)

type gitHubPresenter struct {
	repo        *gist7480523.GoPackageRepo
	gitHubOwner string
	gitHubRepo  string

	comparison *githubComparison
}

func newGitHubPresenter(repo *gist7480523.GoPackageRepo, gitHubOwner, gitHubRepo string) Presenter {
	goPackage := repo.GoPackages()[0]
	comparison := newGithubComparison(gitHubOwner, gitHubRepo, goPackage.Dir.Repo.VcsLocal, goPackage.Dir.Repo.VcsRemote)
	gist7802150.MakeUpdated(comparison)

	p := &gitHubPresenter{repo: repo, gitHubOwner: gitHubOwner, gitHubRepo: gitHubRepo, comparison: comparison}
	return p
}

func (this gitHubPresenter) Repo() *gist7480523.GoPackageRepo {
	return this.repo
}

func (this gitHubPresenter) HomePage() *template.URL {
	switch goPackage := this.repo.GoPackages()[0]; {
	case strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/"):
		url := template.URL("https://github.com/" + this.gitHubOwner + "/" + this.gitHubRepo)
		return &url
	default:
		url := template.URL("http://" + goPackage.Bpkg.ImportPath)
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
	if this.comparison.err != nil {
		return nil
	}
	out := make(chan Change)
	go func() {
		for index := range this.comparison.cc.Commits {
			change := Change{
				Message: firstParagraph(*this.comparison.cc.Commits[len(this.comparison.cc.Commits)-1-index].Commit.Message),
				Url:     template.URL(*this.comparison.cc.Commits[len(this.comparison.cc.Commits)-1-index].HTMLURL),
			}
			if commentCount := this.comparison.cc.Commits[len(this.comparison.cc.Commits)-1-index].Commit.CommentCount; commentCount != nil && *commentCount > 0 {
				change.Comments.Count = *commentCount
				change.Comments.Url = template.URL(*this.comparison.cc.Commits[len(this.comparison.cc.Commits)-1-index].HTMLURL + "#comments")
			}
			out <- change
		}
		close(out)
	}()
	return out
}

// ---

var gh = github.NewClient(nil)

func newGithubComparison(gitHubOwner, gitHubRepo string, local *exp13.VcsLocal, remote *exp13.VcsRemote) *githubComparison {
	this := &githubComparison{gitHubOwner: gitHubOwner, gitHubRepo: gitHubRepo}
	this.AddSources(local, remote)
	return this
}

type githubComparison struct {
	gitHubOwner string
	gitHubRepo  string

	cc  *github.CommitsComparison
	err error

	gist7802150.DepNode2
}

func (this *githubComparison) Update() {
	localRev := this.GetSources()[0].(*exp13.VcsLocal).LocalRev
	remoteRev := this.GetSources()[1].(*exp13.VcsRemote).RemoteRev

	this.cc, _, this.err = gh.Repositories.CompareCommits(this.gitHubOwner, this.gitHubRepo, localRev, remoteRev)
}
