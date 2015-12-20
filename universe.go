package main

import (
	"go/build"
	"html/template"
	"log"
	"sync"

	"github.com/bradfitz/iter"
	"github.com/shurcooL/Go-Package-Store/pkg"
	vcs2 "github.com/shurcooL/go/vcs"
	"golang.org/x/tools/go/vcs"
)

type importPath struct {
	importPath string
}

type importPathRevision struct {
	importPath string
	revision   string
}

func newGoUniverse() *goUniverse {
	u := &goUniverse{
		In:           make(chan importPathRevision, 64),
		InImportPath: make(chan importPath, 64),
		phase2:       make(chan *pkg.Repo, 64),
		Out:          make(chan *pkg.Repo, 64),

		repos: make(map[string]*pkg.Repo),
	}

	for range iter.N(8) {
		u.wg1.Add(1)
		go u.worker() // Phase 1 (i.e., In) to phase 2 worker.
	}
	for range iter.N(8) {
		u.wg1B.Add(1)
		go u.workerB() // Phase 1 (i.e., InImportPath) to phase 2 worker.
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

	InImportPath chan importPath
	wg1B         sync.WaitGroup

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
	close(u.InImportPath)
}

// phase12Cleanup waits for phase 1->2 worker to finish and closes phase2 channel.
func (u *goUniverse) phase12Cleanup() {
	u.wg1.Wait()
	u.wg1B.Wait()
	close(u.phase2)
}

// phase23Cleanup waits for phase 2->3 worker to finish and closes Out channel.
func (u *goUniverse) phase23Cleanup() {
	u.wg2.Wait()
	close(u.Out)
}

// worker for phase 1, sends unique repos to phase 2.
func (u *goUniverse) worker() {
	defer u.wg1.Done()
	for p := range u.In {
		//started := time.Now()
		// Determine repo root.
		// This is potentially somewhat slow.
		rr, err := vcs.RepoRootForImportPath(p.importPath, false)
		if err != nil {
			panic(err) // TODO.
		}
		//fmt.Printf("rr: %v ms.\n", time.Since(started).Seconds()*1000)

		var repo *pkg.Repo
		u.reposMu.Lock()
		if _, ok := u.repos[rr.Root]; !ok {
			repo = &pkg.Repo{
				Root:      rr.Root,
				RemoteURL: rr.Repo,
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

// worker for phase 1, sends unique repos to phase 2.
func (u *goUniverse) workerB() {
	defer u.wg1B.Done()
	for p := range u.InImportPath {
		//started := time.Now()
		// Determine repo root and local revision.
		// This is potentially somewhat slow.
		bpkg, err := build.Import(p.importPath, "", build.FindOnly)
		if err != nil {
			log.Println("build.Import:", err)
			continue
		}
		//goon.DumpExpr(bpkg)
		if bpkg.Goroot {
			continue
		}
		vcs := vcs2.New(bpkg.Dir)
		if vcs == nil {
			log.Println("not in VCS:", bpkg.Dir)
			continue
		}
		repoRoot := vcs.RootPath()[len(bpkg.SrcRoot)+1:] // TODO: Consider sym links, etc.
		//fmt.Printf("build + vcs: %v ms.\n", time.Since(started).Seconds()*1000)

		var repo *pkg.Repo
		u.reposMu.Lock()
		if _, ok := u.repos[repoRoot]; !ok {
			repo = &pkg.Repo{
				Root: repoRoot,
				VCS:  vcs,
				// TODO: Maybe keep track of import paths inside, etc.
			}
			u.repos[repoRoot] = repo
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

// Phase 2 figures out repo remote revision (and local if needed).
func (u *goUniverse) phase2Worker() {
	defer u.wg2.Done()
	for p := range u.phase2 {
		//started := time.Now()
		// Determine remote revision.
		// This is slow because it requires a network operation.
		var remoteVCS vcs2.Remote
		var localVCS vcs2.Vcs
		switch {
		case p.VCS != nil:
			remoteVCS = p.VCS
			localVCS = p.VCS
		case p.RR != nil && p.RR.VCS.Cmd == vcs2.Git.VcsType():
			remoteVCS = vcs2.NewRemote(vcs2.Git, template.URL(p.RR.Repo))
		}
		remoteRevision := remoteVCS.GetRemoteRev()
		//fmt.Printf("remoteVCS.GetRemoteRev: %v ms.\n", time.Since(started).Seconds()*1000)

		p.Remote = pkg.Remote{
			Revision: remoteRevision,
		}

		// TODO: Organize.
		if p.Local.Revision == "" && localVCS != nil {
			p.Local = pkg.Local{
				Revision: localVCS.GetLocalRev(),
			}

			// TODO: Organize.
			p.RemoteURL = localVCS.GetRemote()
		}

		u.Out <- p
	}
}
