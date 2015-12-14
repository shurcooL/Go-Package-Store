// Package util has internal utilities.
package util

import "github.com/shurcooL/go/gists/gist7480523"

// GetRootPath computes and returns the root path of the given goPackage.
func GetRootPath(goPackage *gist7480523.GoPackage) (rootPath string) {
	if goPackage.Bpkg.Goroot {
		return ""
	}

	goPackage.UpdateVcs()
	/*if this.Bpkg.Goroot == false { // Optimization that assume packages under Goroot are not under vcs
		//gist7802150.MakeUpdated(this.Dir)
		if vcs := vcs.New(this.path); vcs != nil {
			reposLock.Lock()
			if repo, ok := repos[vcs.RootPath()]; ok {
				this.Repo = repo
			} else {
				this.Repo = exp13.NewVcsState(vcs)
				repos[vcs.RootPath()] = this.Repo
			}
			reposLock.Unlock()
		}
	}*/

	if goPackage.Dir.Repo == nil {
		return ""
	} else {
		return goPackage.Dir.Repo.Vcs.RootPath()
	}
}
