// Package presenter defines domain types for Go Package Store presenters.
package presenter

import (
	"context"
)

// Presenter returns a Presentation for r, or nil if it can't.
type Presenter func(ctx context.Context, r Repo) *Presentation

// Repo represents a single repository to be presented.
// It contains the input for a Presenter.
type Repo struct {
	// Root is the import path corresponding to the root of the repository.
	Root string

	// RepoURL is the repository URL, including scheme, as determined dynamically from the import path.
	RepoURL string

	LocalRevision  string
	RemoteRevision string
}

// Presentation provides information about a Go package repo with an available update.
// It contains the output of a Presenter.
type Presentation struct {
	HomeURL  string   // Home URL of the Go package. Optional (empty string means none available).
	ImageURL string   // Image representing the Go package, typically its owner.
	Changes  []Change // List of changes, starting with the most recent.
	Error    error    // Any error that occurred during presentation, to be displayed to user.
}

// Change represents a single commit message.
type Change struct {
	Message  string   // Commit message of this change.
	URL      string   // URL of this change.
	Comments Comments // Comments on this change.
}

// Comments represents change discussion.
type Comments struct {
	Count int    // Count of comments on this change.
	URL   string // URL of change discussion. Optional (empty string means none available).
}
