package repo

import (
	"fmt"
	"time"
)

type MockUpdater struct{}

func (MockUpdater) Update(importPathPattern string) error {
	fmt.Println("MockUpdater: got update request:", importPathPattern)
	mockDelay := time.Second
	fmt.Printf("pretending to update (actually sleeping for %v)", mockDelay)
	time.Sleep(mockDelay)
	return nil
}
