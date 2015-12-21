package main

import (
	"encoding/json"
	"os"
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
	return g, err
}
