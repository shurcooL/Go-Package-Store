package main

import (
	"os"

	"github.com/kardianos/govendor/vendorfile"
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
	return v, err
}
