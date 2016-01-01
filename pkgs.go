package main

import (
	"sync"

	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/presenter"
)

type GoPackageList struct {
	// TODO: Merge the List and OrderedList into a single struct to better communicate that it's a single data structure.
	sync.Mutex
	OrderedList []*RepoPresenter          // OrderedList has the same contents as List, but gives it a stable order.
	List        map[string]*RepoPresenter // Map key is repoRoot.
}

type RepoPresenter struct {
	Repo *pkg.Repo
	presenter.Presenter
}
