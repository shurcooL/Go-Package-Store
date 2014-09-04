package presenter

import (
	"html/template"

	"github.com/shurcooL/go/gists/gist7480523"
)

type genericPresenter struct {
	repo *gist7480523.GoPackageRepo
}

func (this genericPresenter) Repo() *gist7480523.GoPackageRepo {
	return this.repo
}
func (_ genericPresenter) HomePage() *template.URL { return nil }
func (_ genericPresenter) Image() template.URL {
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}
func (_ genericPresenter) Changes() <-chan Change { return nil }

// changeMessage is a simple implementation of Change.
type changeMessage string

func (cm changeMessage) Message() string {
	return string(cm)
}
