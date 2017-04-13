package updater

import (
	"fmt"

	"github.com/shurcooL/Go-Package-Store"
)

// Gopath is an Updater that updates Go packages in local GOPATH workspaces.
type Gopath struct{}

// Update specified repository to latest version.
func (Gopath) Update(repo *gps.Repo) error {
	if repo.VCS == nil || repo.Path == "" || repo.Cmd == nil {
		return fmt.Errorf("missing information needed to update Go package in GOPATH: %#v", repo)
	}

	fmt.Printf("cd %s\n", repo.Path)
	fmt.Printf("%s %s", repo.Cmd.Cmd, repo.Cmd.DownloadCmd)
	err := repo.Cmd.Download(repo.Path)
	return err
}
