package repo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shurcooL/Go-Package-Store/pkg"
)

// NewGitSubrepoUpdater returns an Updater that updates Go packages vendored using git-subrepo.
// dir controls where the `git-subrepo` binary is executed. If empty string, current working
// directory is used. If `git-subrepo` binary is not available in PATH, an error will be returned.
func NewGitSubrepoUpdater(dir string) (Updater, error) {
	if _, err := exec.LookPath("git-subrepo"); err != nil {
		return nil, fmt.Errorf("git-subrepo binary is required for updating, but not available: %v", err)
	}
	return gitSubrepoUpdater{dir: dir}, nil
}

// gitSubrepoUpdater is an Updater that updates Go packages listed in vendor.json.
type gitSubrepoUpdater struct {
	dir string // Where to execute `git-subrepo` binary.
}

func (u gitSubrepoUpdater) Update(repo *pkg.Repo) error {
	/*
		```
		git subrepo pull vendor/github.com/golang/glog
		```

		Maybe use --branch=<branch-name> option to specify commit.
	*/

	cmd := exec.Command("echo", "git-subrepo", "something", repo.ImportPathPattern()+"@"+repo.Remote.Revision)
	fmt.Print(strings.Join(cmd.Args, " "))
	cmd.Dir = u.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
