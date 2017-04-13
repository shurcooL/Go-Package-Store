package updater

import (
	"fmt"
	"time"

	"github.com/shurcooL/Go-Package-Store"
)

// Mock is a mock updater.
type Mock struct{}

// Update pretends to update specified repository to latest version.
func (Mock) Update(repo *gps.Repo) error {
	fmt.Println("Mock: got update request:", repo.Root)
	const mockDelay = 3 * time.Second
	fmt.Printf("pretending to update (actually sleeping for %v)", mockDelay)
	time.Sleep(mockDelay)
	return nil
}
