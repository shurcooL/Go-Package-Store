package main

import (
	"log"
	"os"

	"github.com/kardianos/govendor/vendorfile"
	"github.com/shurcooL/go/exp/14"
	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/shurcooL/go/gists/gist7802150"
)

// readGovendor reads a vendor.json file at path.
func readGovendor(path string) (vendorfile.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return vendorfile.File{}, err
	}
	defer f.Close()

	var v vendorfile.File
	err = v.Unmarshal(f)
	if err != nil {
		return vendorfile.File{}, err
	}

	return v, nil
}

// goPackagesFromGovendor implements exp14.GoPackageList, but sources
// the list of Go packages and their current revisions from the vendor.json file at path.
type goPackagesFromGovendor struct {
	path string

	Entries []*gist7480523.GoPackage

	gist7802150.DepNode2
}

func newGoPackagesFromGovendor(path string) exp14.GoPackageList {
	return &goPackagesFromGovendor{path: path}
}

func (this *goPackagesFromGovendor) Update() {
	// TODO: Have a source?

	v, err := readGovendor(this.path)
	if err != nil {
		log.Fatalln("readGovendor:", err)
	}

	this.Entries = nil
	for _, dependency := range v.Package {
		goPackage := gist7480523.GoPackageFromImportPath(dependency.Path)
		if goPackage == nil {
			log.Printf("warning: Govendor dependency %q not found in your GOPATH, skipping\n", dependency.Path)
			continue
		}

		// TODO: Can try to optimize by not blocking on UpdateVcs() here...
		gist7802150.MakeUpdatedLock.Unlock() // HACK: Needed because UpdateVcs() calls MakeUpdated().
		goPackage.UpdateVcs()
		gist7802150.MakeUpdatedLock.Lock() // HACK
		if goPackage.Dir.Repo == nil {
			continue
		}
		goPackage.Dir.Repo.Vcs = &fixedLocalRevVcs{LocalRev: dependency.Revision, Vcs: goPackage.Dir.Repo.Vcs}

		this.Entries = append(this.Entries, goPackage)
	}
}

func (this *goPackagesFromGovendor) List() []*gist7480523.GoPackage {
	return this.Entries
}
