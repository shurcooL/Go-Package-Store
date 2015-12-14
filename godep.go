package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/shurcooL/go/exp/14"
	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/shurcooL/go/gists/gist7802150"
	"github.com/shurcooL/go/vcs"
)

// Godeps describes what a package needs to be rebuilt reproducibly.
// It's the same information stored in file Godeps.
type Godeps struct {
	ImportPath string
	GoVersion  string
	Packages   []string `json:",omitempty"` // Arguments to save, if any.
	Deps       []Dependency
}

// A Dependency is a specific revision of a package.
type Dependency struct {
	ImportPath string
	Comment    string `json:",omitempty"` // Description of commit, if present.
	Rev        string // VCS-specific commit ID.
}

// readGodeps reads a Godeps.json file at path.
func readGodeps(path string) (Godeps, error) {
	f, err := os.Open(path)
	if err != nil {
		return Godeps{}, err
	}
	defer f.Close()

	var g Godeps
	err = json.NewDecoder(f).Decode(&g)
	if err != nil {
		return Godeps{}, err
	}

	return g, nil
}

// ---

// goPackagesFromGodeps implements exp14.GoPackageList, but sources
// the list of Go packages and their current revisions from the Godeps.json file at path.
type goPackagesFromGodeps struct {
	path string

	Entries []*gist7480523.GoPackage

	gist7802150.DepNode2
}

func newGoPackagesFromGodeps(path string) exp14.GoPackageList {
	return &goPackagesFromGodeps{path: path}
}

func (this *goPackagesFromGodeps) Update() {
	// TODO: Have a source?

	g, err := readGodeps(this.path)
	if err != nil {
		log.Fatalln("readGodeps:", err)
	}

	this.Entries = nil
	for _, dependency := range g.Deps {
		goPackage := gist7480523.GoPackageFromImportPath(dependency.ImportPath)
		if goPackage == nil {
			// TODO: Improve this; don't use local GOPATH for remote vcs. Use vcs.NewRemote().
			log.Printf("warning: Godeps dependency %q not found in your GOPATH, skipping\n", dependency.ImportPath)
			continue
		}

		// TODO: Can try to optimize by not blocking on UpdateVcs() here...
		gist7802150.MakeUpdatedLock.Unlock() // HACK: Needed because UpdateVcs() calls MakeUpdated().
		goPackage.UpdateVcs()
		gist7802150.MakeUpdatedLock.Lock() // HACK
		if goPackage.Dir.Repo == nil {
			continue
		}
		goPackage.Dir.Repo.Vcs = &fixedLocalRevVcs{LocalRev: dependency.Rev, Vcs: goPackage.Dir.Repo.Vcs}

		this.Entries = append(this.Entries, goPackage)
	}
}

func (this *goPackagesFromGodeps) List() []*gist7480523.GoPackage {
	return this.Entries
}

// fixedLocalRevVcs represents a virtual VCS with the specified LocalRev,
// clean working directory, default branch checked out.
type fixedLocalRevVcs struct {
	vcs.Vcs

	LocalRev string
}

func (f *fixedLocalRevVcs) GetLocalRev() string {
	return f.LocalRev
}

func (f *fixedLocalRevVcs) IsContained(rev string) bool {
	// This is needed so that it consideres all different remote versions as updates (instead of needing to push).
	return false
}

func (f *fixedLocalRevVcs) GetStatus() string {
	// Clean working directory.
	return ""
}

func (f *fixedLocalRevVcs) GetLocalBranch() string {
	// Default branch checked out.
	return f.Vcs.GetDefaultBranch()
}
