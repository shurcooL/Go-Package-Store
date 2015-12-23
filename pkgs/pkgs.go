package pkgs

import (
	"sync"

	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/presenter"
)

type GoPackageList struct {
	sync.Mutex
	List map[string]*RepoPresenter // Map key is repoRoot.
}

type RepoPresenter struct {
	Repo      *pkg.Repo
	Presenter presenter.Presenter
}
