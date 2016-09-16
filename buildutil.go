package main

import (
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/vcs"
)

// forEachRepository calls found for each repository it finds in all GOPATH workspaces.
func forEachRepository(found func(localRepo)) {
	for _, workspace := range filepath.SplitList(build.Default.GOPATH) {
		srcRoot := filepath.Join(workspace, "src")
		if _, err := os.Stat(srcRoot); os.IsNotExist(err) {
			continue
		}
		// TODO: Confirm that ignoring filepath.Walk error is correct/desired behavior.
		_ = filepath.Walk(srcRoot, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				log.Printf("can't stat file %s: %v\n", path, err)
				return nil
			}
			if !fi.IsDir() {
				return nil
			}
			if strings.HasPrefix(fi.Name(), ".") || strings.HasPrefix(fi.Name(), "_") || fi.Name() == "testdata" {
				return filepath.SkipDir
			}
			// Determine repo root. This is potentially somewhat slow.
			vcsCmd, root, err := vcs.FromDir(path, srcRoot)
			if err != nil {
				// Directory not under VCS.
				return nil
			}
			found(localRepo{Path: path, Root: root, VCS: vcsCmd})
			return filepath.SkipDir // No need to descend inside repositories.
		})
	}
}
