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
func (this genericPresenter) HomePage() *template.URL {
	url := template.URL("http://" + this.repo.GoPackages()[0].Bpkg.ImportPath)
	return &url
}
func (_ genericPresenter) Image() template.URL {
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}
func (_ genericPresenter) Changes() <-chan Change { return nil }
