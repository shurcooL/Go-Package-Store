package main

import (
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/bradfitz/iter"
	"github.com/shurcooL/Go-Package-Store/pkg"
	vcs2 "github.com/shurcooL/go/vcs"
	"golang.org/x/tools/go/vcs"
)

type importPathRevision struct {
	importPath string
	revision   string
}

func newGoUniverse() *goUniverse {
	u := &goUniverse{
		In:     make(chan importPathRevision, 64),
		phase2: make(chan *pkg.Repo, 64),
		Out:    make(chan *pkg.Repo, 64),

		repos: make(map[string]*pkg.Repo),
	}
	for range iter.N(8) {
		u.wg1.Add(1)
		go u.worker() // Phase 1 (i.e., In) to phase 2 worker.
	}
	go u.phase12Cleanup()
	for range iter.N(8) {
		u.wg2.Add(1)
		go u.phase2Worker() // Phase 2 to phase 3 (i.e., Out) worker.
	}
	go u.phase23Cleanup()
	return u
}

// TODO: Rename to goWorkspace or something. It's a local workspace environment, meaning each repo has local and remote components.
type goUniverse struct {
	In  chan importPathRevision
	wg1 sync.WaitGroup

	phase2 chan *pkg.Repo
	wg2    sync.WaitGroup

	// Out is the output of processed repos (complete with local and remote revisions).
	Out chan *pkg.Repo

	reposMu sync.Mutex
	repos   map[string]*pkg.Repo // Map key is repoRoot.
}

// Done should be called after In is completely populated.
func (u *goUniverse) Done() {
	close(u.In)
}

// phase12Cleanup waits for phase 1->2 worker to finish and closes phase2 channel.
func (u *goUniverse) phase12Cleanup() {
	u.wg1.Wait()
	close(u.phase2)
}

// phase23Cleanup waits for phase 2->3 worker to finish and closes Out channel.
func (u *goUniverse) phase23Cleanup() {
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
		started := time.Now()
		// Determine repo root.
		// This is potentially somewhat slow.
		rr, err := vcs.RepoRootForImportPath(p.importPath, false)
		if err != nil {
			panic(err) // TODO.
		}
		fmt.Printf("rr: %v ms.\n", time.Since(started).Seconds()*1000)
		total += time.Since(started).Seconds() * 1000

		var repo *pkg.Repo
		u.reposMu.Lock()
		if _, ok := u.repos[rr.Root]; !ok {
			repo = &pkg.Repo{
				Root: rr.Root,
				Local: pkg.Local{
					Revision: p.revision,
				},
				RR: rr,
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
		started := time.Now()
		// Determine remote revision.
		// This is slow because it requires a network operation.
		var revision string
		switch p.RR.VCS.Cmd {
		case vcs2.Git.VcsType():
			vcs := vcs2.NewRemote(vcs2.Git, template.URL(p.RR.Repo))
			revision = vcs.GetRemoteRev()
		}
		fmt.Printf("vcs.GetRemoteRev: %v ms.\n", time.Since(started).Seconds()*1000)

		p.Remote = pkg.Remote{
			Revision: revision,
		}

		u.Out <- p
	}
}
