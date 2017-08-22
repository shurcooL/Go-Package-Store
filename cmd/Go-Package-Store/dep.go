package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// depDir ensures that "Gopkg.toml" file exists at path,
// and returns the directory that contains it.
func depDir(path string) (string, error) {
	// Check that "Gopkg.toml" file exists.
	if fi, err := os.Stat(path); err != nil {
		return "", err
	} else if !(fi.Name() == "Gopkg.toml" && !fi.IsDir()) {
		return "", fmt.Errorf("%v is not a Gopkg.toml file", path)
	}
	dir := filepath.Dir(path) // Directory containing the Gopkg.toml file.
	return dir, nil
}

// runDepStatus runs dep status in directory dir,
// returning a parsed list of dependencies and their status.
func runDepStatus(dir string) (depDependencies, error) {
	cmd := exec.Command("dep", "status", "-json")
	cmd.Dir = dir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	var dependencies depDependencies
	err = json.NewDecoder(stdout).Decode(&dependencies)
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	return dependencies, err
}

type depDependencies []depDependency

type depDependency struct {
	ProjectRoot string // E.g., "github.com/google/go-github".
	Revision    string // E.g., "6afafa88c26eb51b33a8307c944bd2f0ef227af7".
	Latest      string // E.g., "fe4b6036cb400908b8903d3df9444b9599a60d57".
}
