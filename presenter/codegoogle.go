package presenter

import (
	"html/template"
	"net/url"
	"strings"

	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/sourcegraph/go-vcs/vcs"
	"github.com/sourcegraph/vcsstore/vcsclient"
)

type codeGooglePresenter struct {
	repo *gist7480523.GoPackageRepo

	comparison codeGoogleComparison
}

func newCodeGooglePresenter(repo *gist7480523.GoPackageRepo) Presenter {
	return &codeGooglePresenter{
		repo:       repo,
		comparison: newCodeGoogleComparison(repo),
	}
}

func (this codeGooglePresenter) Repo() *gist7480523.GoPackageRepo {
	return this.repo
}
func (this codeGooglePresenter) HomePage() *template.URL {
	url := template.URL("https://" + repo.RepoImportPath())
	return &url
}
func (_ codeGooglePresenter) Image() template.URL {
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}
func (this codeGooglePresenter) Changes() <-chan Change {
	if this.comparison.err != nil {
		return nil
	}
	out := make(chan Change)
	go func() {
		foundLocalRev := false
		for _, commit := range this.comparison.commits {
			// Break out when/if we reach the current local revision.
			if commit.ID == vcs.CommitID(this.repo.GoPackages()[0].Dir.Repo.VcsLocal.LocalRev) {
				foundLocalRev = true
				break
			}
			out <- changeMessage(firstParagraph(commit.Message))
		}
		if !foundLocalRev {
			out <- changeMessage("... (there may be more changes, not shown)")
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

// ---

var sg = vcsclient.New(&url.URL{Scheme: "http", Host: "vcsstore.sourcegraph.com"}, nil)

type codeGoogleComparison struct {
	commits []*vcs.Commit
	err     error
}

func newCodeGoogleComparison(repo *gist7480523.GoPackageRepo) (c codeGoogleComparison) {
	cloneUrl, err := url.Parse("https://" + repo.RepoImportPath())
	if err != nil {
		c.err = err
		return
	}

	r, err := sg.Repository("hg", cloneUrl) // code.google.com/p/... repos are known to use Mercurial.
	if err != nil {
		c.err = err
		return
	}

	c.commits, _, c.err = r.Commits(vcs.CommitsOptions{
		Head: vcs.CommitID(repo.GoPackages()[0].Dir.Repo.VcsRemote.RemoteRev),
		N:    20, // Cap for now. TODO: Support arbtirary second revision to go until.
	})
	return
}
