package repo

import (
	"fmt"
	"time"

	"github.com/shurcooL/Go-Package-Store/pkgs"
)

type MockUpdater struct {
	GoPackages *pkgs.GoPackageList
}

func (u MockUpdater) Update(importPathPattern string) error {
	fmt.Println("MockUpdater: got update request:", importPathPattern)

	root := importPathPattern[:len(importPathPattern)-4]

	u.GoPackages.Lock()
	_, ok := u.GoPackages.List[root]
	u.GoPackages.Unlock()

	if !ok {
		return fmt.Errorf("import path pattern %q not found in GOPATH", importPathPattern)
	}

	mockDelay := time.Second
	fmt.Printf("pretending to update (actually sleeping for %v)", mockDelay)
	time.Sleep(mockDelay)

	// Delete repo from list.
	u.GoPackages.Lock()
	// TODO: Consider marking the repo as "Updated" and display it that way, etc.
	for i := range u.GoPackages.OrderedList {
		if u.GoPackages.OrderedList[i].Repo.Root == root {
			u.GoPackages.OrderedList = append(u.GoPackages.OrderedList[:i], u.GoPackages.OrderedList[i+1:]...)
			break
		}
	}
	delete(u.GoPackages.List, root)
	u.GoPackages.Unlock()
	return nil
}
