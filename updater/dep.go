package updater

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shurcooL/Go-Package-Store"
)

// Dep is an Updater that updates Go packages in a project managed by dep.
//
// It requires the dep binary to be available in PATH.
type Dep struct {
	// Dir specifies where the dep binary is executed.
	// If empty, current working directory is used.
	Dir string
}

// Update specified repository to latest version by calling
// "dep ensure -update <repo-root>" in d.Dir directory.
func (d Dep) Update(repo *gps.Repo) error {
	cmd := exec.Command("dep", "ensure", "-update", repo.Root)
	fmt.Println(strings.Join(cmd.Args, " "))
	cmd.Dir = d.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
