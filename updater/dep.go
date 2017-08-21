package updater

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shurcooL/Go-Package-Store"
)

// NewDep returns an Updater that updates Go packages in a project managed by dep.
// dir controls where the dep binary is executed. If empty string, current working
// directory is used. If dep binary is not available in PATH, an error will be returned.
func NewDep(dir string) (gps.Updater, error) {
	if _, err := exec.LookPath("dep"); err != nil {
		return nil, fmt.Errorf("dep binary is required for updating, but not available: %v", err)
	}
	return dep{dir: dir}, nil
}

// dep is an Updater that updates Go packages in a project managed by dep.
type dep struct {
	dir string // Directory where to execute dep binary.
}

func (gu dep) Update(repo *gps.Repo) error {
	cmd := exec.Command("dep", "ensure", "-update", repo.Root)
	fmt.Println(strings.Join(cmd.Args, " "))
	cmd.Dir = gu.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
