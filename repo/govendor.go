package repo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shurcooL/Go-Package-Store/pkg"
)

// NewGovendorUpdater returns an Updater that updates Go packages listed in vendor.json.
// dir controls where the `govendor` binary is executed. If empty string, current working
// directory is used. If `govendor` binary is not available in PATH, an error will be returned.
func NewGovendorUpdater(dir string) (Updater, error) {
	if _, err := exec.LookPath("govendor"); err != nil {
		return nil, fmt.Errorf("govendor binary is required for updating, but not available: %v", err)
	}
	return govendorUpdater{dir: dir}, nil
}

// govendorUpdater is an Updater that updates Go packages listed in vendor.json.
type govendorUpdater struct {
	dir string // Where to execute `govendor` binary.
}

func (gu govendorUpdater) Update(repo *pkg.Repo) error {
	cmd := exec.Command("govendor", "fetch", repo.ImportPathPattern()+"@"+repo.Remote.Revision)
	fmt.Print(strings.Join(cmd.Args, " "))
	cmd.Dir = gu.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
