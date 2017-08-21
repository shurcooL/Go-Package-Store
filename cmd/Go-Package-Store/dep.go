package main

import (
	"github.com/BurntSushi/toml"
)

type depLock struct {
	Projects []depLockedProject `toml:"projects"`
}

type depLockedProject struct {
	Name     string `toml:"name"`
	Revision string `toml:"revision"`
}

// readDepLock reads a Gopkg.lock file at path.
func readDepLock(path string) (depLock, error) {
	var l depLock
	_, err := toml.DecodeFile(path, &l)
	return l, err
}
