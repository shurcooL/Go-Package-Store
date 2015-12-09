package repo

import "fmt"

type GopathUpdater struct{}

func (GopathUpdater) Update(importPathPattern string) error {
	return fmt.Errorf("not implemented")
}
