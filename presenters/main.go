package presenter

import (
	"html/template"
	"strings"

	"github.com/google/go-github/github"
	"github.com/shurcooL/go/exp/13"
	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/shurcooL/go/gists/gist7802150"
)

type changeProvider func(repo *gist7480523.GoPackageRepo) Change

var changeProviders []changeProvider

func addProvider(p changeProvider) {
	changeProviders = append(changeProviders, p)
}

type Change interface {
	Repo() *gist7480523.GoPackageRepo

	WebLink() *template.URL
	AvatarUrl() template.URL
	Changes() <-chan github.RepositoryCommit

	Comparison() *GithubComparison
}

func New(repo *gist7480523.GoPackageRepo) Change {
	// TODO: Try to figure out vcs provider with a more constant-time operation.
	// TODO: Potentially check in parallel.
	for _, provider := range changeProviders {
		if presenter := provider(repo); presenter != nil {
			return presenter
		}
	}

	return genericPresenter{repo: repo}
}

/*func init() {
	// GitHub
	addProvider(func(repo *gist7480523.GoPackageRepo) Change {
		goPackage := repo.GoPackages()[0]
		if strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/") {
			return NewGitHubChangePresenter(goPackage)
		}
		return nil
	})

	// gopkg.in
	addProvider(func(repo *gist7480523.GoPackageRepo) Change {
		goPackage := repo.GoPackages()[0]
		if strings.HasPrefix(goPackage.Bpkg.ImportPath, "gopkg.in/") {
			return NewGopkgInChangePresenter(goPackage)
		}
		return nil
	})

	// TODO: code.google.com?
}*/

// =====

type genericPresenter struct {
	repo *gist7480523.GoPackageRepo
}

func (this genericPresenter) Repo() *gist7480523.GoPackageRepo {
	return this.repo
}
func (_ genericPresenter) WebLink() *template.URL { return nil }
func (_ genericPresenter) AvatarUrl() template.URL {
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}
func (_ genericPresenter) Changes() <-chan github.RepositoryCommit { return nil }
func (_ genericPresenter) Comparison() *GithubComparison           { return nil }

// =====

type gitHubChangePresenter struct {
	repo *gist7480523.GoPackageRepo
}

func NewGitHubChangePresenter(repo *gist7480523.GoPackageRepo) Change {
	p := &gitHubChangePresenter{repo: repo}
	return p
}

func (this gitHubChangePresenter) Repo() *gist7480523.GoPackageRepo {
	return this.repo
}

func (this gitHubChangePresenter) WebLink() *template.URL {
	goPackage := this.repo.GoPackages()[0]

	// TODO: Factor these out into a nice interface...
	switch {
	case strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/"):
		importPathElements := strings.Split(goPackage.Bpkg.ImportPath, "/")
		url := template.URL("https://github.com/" + importPathElements[1] + "/" + importPathElements[2])
		return &url
	case strings.HasPrefix(goPackage.Bpkg.ImportPath, "gopkg.in/"):
		// TODO
		return nil
	case strings.HasPrefix(goPackage.Dir.Repo.VcsLocal.Remote, "https://github.com/"):
		url := template.URL(strings.TrimSuffix(goPackage.Dir.Repo.VcsLocal.Remote, ".git"))
		return &url
	default:
		return nil
	}
}

func (this gitHubChangePresenter) AvatarUrl() template.URL {
	// Use the repo owner avatar image.
	if this.Comparison() != nil {
		if user, _, err := gh.Users.Get(this.Comparison().owner); err == nil && user.AvatarURL != nil {
			return template.URL(*user.AvatarURL)
		}
	}
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}

func (this gitHubChangePresenter) Comparison() *GithubComparison {
	// TODO
	return nil
}

// List of changes, starting with the most recent.
// Precondition is that this.Comparison != nil.
func (this gitHubChangePresenter) Changes() <-chan github.RepositoryCommit {
	out := make(chan github.RepositoryCommit)
	go func() {
		for index := range this.Comparison().cc.Commits {
			out <- this.Comparison().cc.Commits[len(this.Comparison().cc.Commits)-1-index]
		}
		close(out)
	}()
	return out
}

// =====

var gh = github.NewClient(nil)

type GithubComparison struct {
	importPath string
	owner      string

	cc  *github.CommitsComparison
	err error

	gist7802150.DepNode2
}

func (this *GithubComparison) Update() {
	localRev := this.GetSources()[0].(*exp13.VcsLocal).LocalRev
	remoteRev := this.GetSources()[1].(*exp13.VcsRemote).RemoteRev

	importPathElements := strings.Split(this.importPath, "/")
	this.cc, _, this.err = gh.Repositories.CompareCommits(importPathElements[1], importPathElements[2], localRev, remoteRev)

	// TODO: Do this better (in the right place, etc.).
	this.owner = importPathElements[1]

	//goon.DumpExpr("GithubComparison.Update()", this.importPath, localRev, remoteRev)
	//fmt.Println(this.err)
}

func NewGithubComparison(importPath string, local *exp13.VcsLocal, remote *exp13.VcsRemote) *GithubComparison {
	this := &GithubComparison{importPath: importPath}
	this.AddSources(local, remote)
	return this
}

// rootPath -> *VcsState
var githubComparisons = make(map[string]*GithubComparison)

// =====

type gopkgInChangePresenter struct {
	Change
}

func NewGopkgInChangePresenter(repo *gist7480523.GoPackageRepo) Change {
	return NewGitHubChangePresenter(repo)
}
