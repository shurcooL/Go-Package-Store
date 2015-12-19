package pkg

import "golang.org/x/tools/go/vcs"

type Repo struct {
	// Root is the import path corresponding to the root of the repository.
	Root string

	// TODO: Consider. Needed for RR.VCS for phase2.
	//       If this is kept, then should remove Root above since it's in here too.
	RR *vcs.RepoRoot

	Local  Local
	Remote Remote
}

type Local struct {
	Revision string
}

type Remote struct {
	Revision string
	//IsContained bool // True if remote commit is contained in the default local branch.
}

// RepoImportPath returns what would be the import path of the root folder of the repository. It may or may not
// be an actual Go package. E.g.,
//
// 	"github.com/owner/repo"
func (r Repo) RepoImportPath() string {
	return r.Root
}

// ImportPathPattern returns an import path pattern that matches all of the Go packages in this repo.
// E.g.,
//
// 	"github.com/owner/repo/..."
func (r Repo) ImportPathPattern() string {
	return r.Root + "/..."
}

// ImportPaths returns a newline separated list of all import paths.
func (r Repo) ImportPaths() string {
	return "ImportPaths not impl"
}
