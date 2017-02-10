package gps

// Presentation provides infomation about a Go package repo with an available update.
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

// Presenter returns a Presentation for r, or nil if it can't.
type Presenter func(r *Repo) *Presentation
