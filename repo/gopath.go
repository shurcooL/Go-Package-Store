package repo

import (
	"fmt"

	"github.com/shurcooL/Go-Package-Store/pkgs"
	"github.com/shurcooL/go-goon"
)

// GopathUpdater is an Updater that updates Go packages in local GOPATH workspaces.
// Those packages are tracked in provided GoPackages.
type GopathUpdater struct {
	// GoPackages is a cached list of Go packages to work with.
	GoPackages *pkgs.GoPackageList
}

func (u GopathUpdater) Update(importPathPattern string) error {
	repoRoot := importPathPattern[:len(importPathPattern)-4]
	goon.DumpExpr(repoRoot)

	u.GoPackages.Lock()
	goPackage, ok := u.GoPackages.List[repoRoot]
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

	rootPath := goPackage.Repo.VCS.RootPath()
	goon.DumpExpr(rootPath)

	vcs := goPackage.Repo.Cmd
	fmt.Printf("cd %s\n", rootPath)
	fmt.Printf("%s %s", vcs.Cmd, vcs.DownloadCmd)
	err := vcs.Download(rootPath)
	return err
}
