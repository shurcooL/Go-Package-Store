package gps

import (
	"github.com/shurcooL/vcsstate"
	"golang.org/x/tools/go/vcs"
)

// Repo represents the state of a single repository.
type Repo struct {
	// Root is the import path corresponding to the root of the repository.
	Root string

	// Exactly one of VCS or RemoteVCS should be not nil.
	// TODO: Consider if it'd be better to split this into two distinct structs.

	// VCS allows getting the state of the VCS.
	VCS vcsstate.VCS
	// Path is the local filesystem path to the repository.
	// It must be set if VCS is not nil.
	Path string
	// Cmd can be used to update this local repository inside a GOPATH workspace.
	// It must be set if VCS is not nil.
	Cmd *vcs.Cmd

	// RemoteVCS allows getting the remote state of the VCS.
	RemoteVCS vcsstate.RemoteVCS
	// RemoteURL is the remote URL, including scheme.
	// It must be set if RemoteVCS is not nil.
	RemoteURL string

	Local struct {
		// RemoteURL is the remote URL, including scheme.
		RemoteURL string

		Revision string
	}
	Remote struct {
		// RepoURL is the repository URL, including scheme, as determined dynamically from the import path.
		RepoURL string

		Branch   string // Default branch, as determined from remote.
		Revision string
	}

	// TODO: Right now, all presenters use Remote.RepoURL as the canonical remote URL, so it must be
	//       always set. This is a little confusing and redundant (since there's also Local.RemoteURL
	//       and just RemoteURL). Should change it so there's only one canonical remote URL for
	//       presenters to use.
}

// ImportPathPattern returns an import path pattern that matches all of the Go packages in this repo.
// E.g.:
//
// 	"github.com/owner/repo/..."
func (r Repo) ImportPathPattern() string {
	return r.Root + "/..."
}
