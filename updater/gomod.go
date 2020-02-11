package updater

import (
	"fmt"
	"os/exec"

	gps "github.com/shurcooL/Go-Package-Store"
	"golang.org/x/xerrors"
)

type GoMod struct{}

func (GoMod) Update(m *gps.Repo) error {
	err := exec.Command("go", "mod", "edit", "-require="+m.Root+"@"+m.Remote.Revision).Run()
	if ee := (*exec.ExitError)(nil); xerrors.As(err, &ee) {
		err = fmt.Errorf("go command exited unsuccessfully: %v\n%s", ee.ProcessState.String(), ee.Stderr)
	}
	return err
}
