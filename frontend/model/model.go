// Package model is a frontend data model for updates.
package model

// RepoPresentation represents a repository update presentation.
//
// TODO: Dedup with workspace.RepoPresentation. Maybe.
type RepoPresentation struct {
	RepoRoot          string
	ImportPathPattern string
	LocalRevision     string
	RemoteRevision    string
	HomeURL           string
	ImageURL          string
	Changes           []Change // TODO: Consider []*Change.
	Error             string

	UpdateState UpdateState

	// TODO: Find a place for this.
	UpdateSupported bool
}

// UpdateState represents the state of an update.
type UpdateState uint8

const (
	// Available represents an available update.
	Available UpdateState = iota

	// Updating represents an update in progress.
	Updating

	// Updated represents a completed update.
	Updated
)

// Change represents a single commit message.
type Change struct {
	Message  string   // Commit message of this change.
	URL      string   // URL of this change.
	Comments Comments // Comments on this change.
}

// Comments represents a change discussion.
//
// TODO: Consider inlining this into Change, we'll see.
type Comments struct {
	Count int
	URL   string
}
