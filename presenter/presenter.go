package presenter

import (
	"html/template"
	"strings"

	"github.com/shurcooL/go/gists/gist7480523"
)

// Presenter is for displaying various info about a given Go package repo with an update available.
type Presenter interface {
	Repo() *gist7480523.GoPackageRepo

	HomePage() *template.URL // Home page url of the Go package, optional (nil means none available).
	Image() template.URL     // Image representing the Go package, typically its owner.
	Changes() <-chan Change  // List of changes, starting with the most recent.
}

// Change represents a single commit message.
type Change interface {
	Message() string
	Url() template.URL
}

type presenterProvider func(repo *gist7480523.GoPackageRepo) Presenter

var presenterProviders []presenterProvider

func addProvider(p presenterProvider) {
	presenterProviders = append(presenterProviders, p)
}

// New takes a repository containing 1 or more Go packages, and returns a Presenter
// for it. It tries to find the best Presenter for the given repository, but falls back
// to a generic presenter if there's nothing better.
func New(repo *gist7480523.GoPackageRepo) Presenter {
	// TODO: Potentially check in parallel.
	for _, provider := range presenterProviders {
		if presenter := provider(repo); presenter != nil {
			return presenter
		}
	}

	return genericPresenter{repo: repo}
}

func init() {
	// GitHub.
	addProvider(func(repo *gist7480523.GoPackageRepo) Presenter {
		switch goPackage := repo.GoPackages()[0]; {
		case strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/"):
			importPathElements := strings.Split(goPackage.Bpkg.ImportPath, "/")
			return newGitHubPresenter(repo, importPathElements[1], importPathElements[2])
		// gopkg.in package.
		case strings.HasPrefix(goPackage.Bpkg.ImportPath, "gopkg.in/"):
			gitHubOwner, gitHubRepo, err := gopkgInImportPathToGitHub(goPackage.Bpkg.ImportPath)
			if err != nil {
				return nil
			}
			return newGitHubPresenter(repo, gitHubOwner, gitHubRepo)
		// Underlying GitHub remote.
		case strings.HasPrefix(goPackage.Dir.Repo.VcsLocal.Remote, "https://github.com/"):
			importPathElements := strings.Split(strings.TrimSuffix(goPackage.Dir.Repo.VcsLocal.Remote[len("https://"):], ".git"), "/")
			return newGitHubPresenter(repo, importPathElements[1], importPathElements[2])
		}
		return nil
	})

	// code.google.com.
	addProvider(func(repo *gist7480523.GoPackageRepo) Presenter {
		goPackage := repo.GoPackages()[0]
		if strings.HasPrefix(goPackage.Bpkg.ImportPath, "code.google.com/p/") {
			return newCodeGooglePresenter(repo)
		}
		return nil
	})
}
