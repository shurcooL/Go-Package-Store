package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/shurcooL/Go-Package-Store/pkg"
	"golang.org/x/tools/go/vcs"
)

type importPathRevision struct {
	importPath string
	revision   string
}

func newGoUniverse() *goUniverse {
	u := &goUniverse{
		In:  make(chan importPathRevision, 64),
		Out: make(chan *pkg.Repo, 64),

		phase2: make(chan *pkg.Repo, 64),
		repos:  make(map[string]*pkg.Repo),
	}
	// TODO: Multiple workers?
	{
		u.wg1.Add(1)
		go u.worker()
	}
	go u.wait1()
	{
		u.wg2.Add(1)
		go u.phase2Worker()
	}
	go u.wait2()
	return u
}

// TODO: Rename to goWorkspace or something. It's a local workspace environment, meaning each repo has local and remote components.
type goUniverse struct {
	In  chan importPathRevision
	wg1 sync.WaitGroup

	phase2 chan *pkg.Repo
	wg2    sync.WaitGroup

	Out chan *pkg.Repo

	reposMu sync.Mutex
	repos   map[string]*pkg.Repo // Map key is repoRoot.
}

func (u *goUniverse) Done() {
	close(u.In)
}

// wait waits for phase 2 to finish and closes Out channel.
func (u *goUniverse) wait1() {
	u.wg1.Wait()
	close(u.phase2)
}

// wait waits for phase 2 to finish and closes Out channel.
func (u *goUniverse) wait2() {
	u.wg2.Wait()
	close(u.Out)
}

var total float64

// worker for phase 1, sends unique repos to phase 2 (TODO: this part).
func (u *goUniverse) worker() {
	defer u.wg1.Done()
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

		var repo *pkg.Repo
		u.reposMu.Lock()
		if _, ok := u.repos[rr.Root]; !ok {
			repo = &pkg.Repo{
				Root: rr.Root,
				Local: &pkg.Local{
					Revision: &p.revision,
				},
				// TODO: Maybe keep track of import paths inside, etc.
			}
			u.repos[rr.Root] = repo
		} else {
			// TODO: Maybe keep track of import paths inside, etc.
		}
		u.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			u.phase2 <- repo
		}
	}
}

// Phase 2 figures out repo remote revision.
func (u *goUniverse) phase2Worker() {
	defer u.wg2.Done()
	for p := range u.phase2 {
		// Determine remote revision.
		// This is slow because it requires a network operation.
		// TODO.
		revision := "TODO"
		p.Remote = &pkg.Remote{
			Revision: &revision,
		}

		u.Out <- p
	}
}