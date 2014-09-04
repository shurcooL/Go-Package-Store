package presenter

import (
	"html/template"

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
}
