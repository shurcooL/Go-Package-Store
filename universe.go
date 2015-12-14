package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/tools/go/vcs"
)

type importPathRevision struct {
	importPath string
	revision   string
}

type rep struct {
	repoRoot string

	local *local
	//remote *exp13.VcsRemote
}

type local struct {
	revision *string
}

func newGoUniverse() *goUniverse {
	u := &goUniverse{
		In: make(chan importPathRevision, 64),

		repos: make(map[string]*rep),
	}
	// TODO: Multiple workers?
	{
		u.wg.Add(1)
		go u.worker()
	}
	return u
}

// TODO: Rename to goWorkspace or something. It's a local workspace environment, meaning each repo has local and remote components.
type goUniverse struct {
	In chan importPathRevision

	wg sync.WaitGroup

	reposMu sync.Mutex
	repos   map[string]*rep // Map key is repoRoot.
}

func (u *goUniverse) Done() {
	close(u.In)
}

// Wait waits for phase 1 (TODO: describe phase 1) to finish.
func (u *goUniverse) Wait() {
	u.wg.Wait()
}

var total float64

// worker for phase 1, sends unique repos to phase 2 (TODO: this part).
func (u *goUniverse) worker() {
	defer u.wg.Done()
	defer func() {
		fmt.Println("total was:", total)
	}()
	for p := range u.In {
		started2 := time.Now()
		// Determine repo root.
		// This is potentially somewhat slow.
		rr, err := vcs.RepoRootForImportPath(p.importPath, false)
		if err != nil {
			panic(err) // TODO.
		}
		fmt.Printf("rr: %v ms.\n", time.Since(started2).Seconds()*1000)
		total += time.Since(started2).Seconds() * 1000

		u.reposMu.Lock()
		if _, ok := u.repos[rr.Root]; !ok {
			u.repos[rr.Root] = &rep{
				repoRoot: rr.Root,
				local: &local{
					revision: &p.revision,
				},
				// TODO: Maybe keep track of import paths inside, etc.
			}
		} else {
			// TODO: Maybe keep track of import paths inside, etc.
		}
		u.reposMu.Unlock()

		// TODO: If new repo, send off to phase 2 channel.
	}
}
