package repo

import (
	"fmt"

	"github.com/shurcooL/Go-Package-Store/internal/util"
	"github.com/shurcooL/go/exp/14"
	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/shurcooL/go/gists/gist7802150"
)

// GopathUpdater is an Updater that updates Go packages in local GOPATH workspaces.
// Those packages are tracked in provided GoPackages.
type GopathUpdater struct {
	// GoPackages is a cached list of Go packages to work with.
	GoPackages exp14.GoPackageList
}

func (u GopathUpdater) Update(importPathPattern string) error {
	// TODO: This uses a legacy gist7802150 caching/cache-invalidation system. It's functional,
	//       but poorly documented, has known flaws (it does not allow concurrent updates),
	//       and very contributor-unfriendly (people don't like packages that have the word "gist" in
	//       the import path, even if's not actually a gist; which is understandable, since it's basically
	//       a package without a name that describes what it's for, something acceptable during rapid
	//       prototyping, but not the finished product). Need to redesign it and replace with
	//       something better.
	//
	//       First step might be to simply drop the caching behavior and hope the user doesn't try
	//       to refresh their browser page very often.

	var updateErr = fmt.Errorf("import path pattern %q not found in GOPATH", importPathPattern)
	gist7802150.MakeUpdated(u.GoPackages)
	for _, goPackage := range u.GoPackages.List() {
		if rootPath := util.GetRootPath(goPackage); rootPath != "" {
			if gist7480523.GetRepoImportPathPattern(rootPath, goPackage.Bpkg.SrcRoot) == importPathPattern {

				vcs := goPackage.Dir.Repo.RepoRoot.VCS
				fmt.Printf("cd %s\n", rootPath)
				fmt.Printf("%s %s", vcs.Cmd, vcs.DownloadCmd)
				updateErr = vcs.Download(rootPath)

				// Invalidate cache of the package's local revision, since it's expected to change after updating.
				gist7802150.ExternallyUpdated(goPackage.Dir.Repo.VcsLocal.GetSources()[1].(gist7802150.DepNode2ManualI))

				break
			}
		}
	}
	return updateErr
}
