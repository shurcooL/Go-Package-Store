// Package repo provides an Updater interface and implementations.
package repo

// Updater is able to update Go packages contained in repositories.
type Updater interface {
	// Update Go packages that match import path pattern to latest version.
	//
	// The only allowed format for import path pattern is "{{.RepoRoot}}/...", where RepoRoot
	// is the import path of repository root (not necessarily a valid Go package).
	// For example, "golang.org/x/net/..." or "github.com/shurcooL/Go-Package-Store/...".
	Update(importPathPattern string) error
}
