package model

// RepoPresentation represents a repository update presentation.
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

type UpdateState uint8

const (
	Available UpdateState = iota
	Updating
	Updated
)

// Change represents a single commit message.
type Change struct {
	Message  string   // Commit message of this change.
	URL      string   // URL of this change.
	Comments Comments // Comments on this change.
}

// Comments represents a change discussion.
// TODO: Consider inlining this into Change, we'll see.
type Comments struct {
	Count int
	URL   string
}
