package presenter

import (
	"html/template"

	"github.com/shurcooL/go/gists/gist7480523"
)

type Presenter interface {
	Repo() *gist7480523.GoPackageRepo

	WebLink() *template.URL
	AvatarUrl() template.URL
	Changes() <-chan Change // List of changes, starting with the most recent.
}

type Change interface {
	Message() string
}
