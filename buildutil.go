package main

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shurcooL/go/trim"
	"github.com/shurcooL/vcsstate"
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

// forEachGitSubrepo calls found for each git subrepo inside vendorDir.
func forEachGitSubrepo(vendorDir string, found func(subrepo)) error {
	remoteVCS, err := vcsstate.NewRemoteVCS(vcs.ByCmd("git"))
	if err != nil {
		return fmt.Errorf("git repos not supported by vcsstate: %v", err)
	}

	err = filepath.Walk(vendorDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}
		if !fi.IsDir() {
			return nil
		}
		if strings.HasPrefix(fi.Name(), ".") {
			return filepath.SkipDir
		}
		remote, commit, err := parseGitRepoFile(path)
		if err != nil {
			return nil
		}
		// Root is the import path corresponding to the root of the repository.
		// It can be determined relative to the vendor directory.
		root, err := filepath.Rel(vendorDir, path)
		if err != nil {
			return err
		}
		found(subrepo{Root: root, RemoteVCS: remoteVCS, RemoteURL: remote, Revision: commit})
		return filepath.SkipDir // No need to descend inside repositories.
	})
	return err
}

// parseGitRepoFile parses a .gitrepo file in directory, returning
// the remote and commit specified within that file.
func parseGitRepoFile(dir string) (remote string, commit string, _ error) {
	remoteBytes, err := exec.Command("git", "config", "--file", filepath.Join(dir, ".gitrepo"), "subrepo.remote").Output()
	if err != nil {
		return "", "", err
	}
	remote = trim.LastNewline(string(remoteBytes))

	commitBytes, err := exec.Command("git", "config", "--file", filepath.Join(dir, ".gitrepo"), "subrepo.commit").Output()
	if err != nil {
		return "", "", err
	}
	commit = trim.LastNewline(string(commitBytes))

	return remote, commit, nil
}
