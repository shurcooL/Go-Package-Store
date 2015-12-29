package repo

import (
	"fmt"

	"github.com/shurcooL/Go-Package-Store/pkgs"
)

// GopathUpdater is an Updater that updates Go packages in local GOPATH workspaces.
// Those packages are tracked in provided GoPackages.
type GopathUpdater struct {
	// GoPackages is a cached list of Go packages to work with.
	GoPackages *pkgs.GoPackageList
}

func (u GopathUpdater) Update(importPathPattern string) error {
	root := importPathPattern[:len(importPathPattern)-4]

	u.GoPackages.Lock()
	goPackage, ok := u.GoPackages.List[root]
	u.GoPackages.Unlock()

	if !ok {
		return fmt.Errorf("import path pattern %q not found in GOPATH", importPathPattern)
	}

	if goPackage.Repo.Cmd == nil {
		return fmt.Errorf("import path pattern %q has goPackage.Repo.Cmd == nil", importPathPattern)
	}

	if goPackage.Repo.VCS == nil {
		return fmt.Errorf("import path pattern %q has goPackage.Repo.VCS == nil", importPathPattern)
	}

	vcs := goPackage.Repo.Cmd
	fmt.Printf("cd %s\n", goPackage.Repo.Path)
	fmt.Printf("%s %s", vcs.Cmd, vcs.DownloadCmd)
	err := vcs.Download(goPackage.Repo.Path)

	if err == nil {
		// Delete repo from list.
		u.GoPackages.Lock()
		// TODO: Consider marking the repo as "Updated" and display it that way, etc.
		for i := range u.GoPackages.OrderedList {
			if u.GoPackages.OrderedList[i].Repo.Root == root {
				u.GoPackages.OrderedList = append(u.GoPackages.OrderedList[:i], u.GoPackages.OrderedList[i+1:]...)
				break
			}
		}
		delete(u.GoPackages.List, root)
		u.GoPackages.Unlock()
	}

	return err
}
