// Package util has internal utilities.
package util

import "github.com/shurcooL/go/gists/gist7480523"

// GetRootPath computes and returns the root path of the given goPackage.
func GetRootPath(goPackage *gist7480523.GoPackage) (rootPath string) {
	if goPackage.Bpkg.Goroot {
		return ""
	}

	goPackage.UpdateVcs()
	if goPackage.Dir.Repo == nil {
		return ""
	} else {
		return goPackage.Dir.Repo.Vcs.RootPath()
	}
}
