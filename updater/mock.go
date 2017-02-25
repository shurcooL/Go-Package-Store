package updater

import (
	"fmt"
	"time"

	"github.com/shurcooL/Go-Package-Store"
)

type Mock struct{}

func (Mock) Update(repo *gps.Repo) error {
	fmt.Println("Mock: got update request:", repo.Root)
	const mockDelay = 3 * time.Second
	fmt.Printf("pretending to update (actually sleeping for %v)", mockDelay)
	time.Sleep(mockDelay)
	return nil
}
