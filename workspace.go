package main

import (
	"fmt"
	"go/build"
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/bradfitz/iter"
	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/pkgs"
	"github.com/shurcooL/Go-Package-Store/presenter"
	vcs2 "github.com/shurcooL/go/vcs"
	"golang.org/x/tools/go/vcs"
)

type importPathRevision struct {
	importPath string
	revision   string
}

// goWorkspace is a workspace environment, meaning each repo has local and remote components.
type goWorkspace struct {
	reposMu sync.Mutex
	repos   map[string]*pkg.Repo // Map key is repoRoot.

	inImportPath chan string
	wg1B         sync.WaitGroup

	inImportPathRevision chan importPathRevision
	wg1A                 sync.WaitGroup

	phase2 chan *pkg.Repo
	wg2    sync.WaitGroup

	// phase3 is the output of processed repos (complete with local and remote revisions),
	// with just enough information to decide if an update should be displayed.
	phase3 chan *pkg.Repo
	wg3    sync.WaitGroup

	// out is the output of processed and presented repos (complete with repo.Presenter).
	out chan *pkgs.RepoPresenter

	newObserver   chan observerRequest
	observers     map[chan *pkgs.RepoPresenter]struct{}
	GoPackageList *pkgs.GoPackageList
}

type observerRequest struct {
	Response chan chan *pkgs.RepoPresenter
}

func NewGoWorkspace() *goWorkspace {
	u := &goWorkspace{
		repos:                make(map[string]*pkg.Repo),
		inImportPath:         make(chan string, 64),
		inImportPathRevision: make(chan importPathRevision, 64),
		phase2:               make(chan *pkg.Repo, 64),
		phase3:               make(chan *pkg.Repo, 64),
		out:                  make(chan *pkgs.RepoPresenter, 64),

		newObserver:   make(chan observerRequest),
		observers:     make(map[chan *pkgs.RepoPresenter]struct{}),
		GoPackageList: &pkgs.GoPackageList{List: make(map[string]*pkgs.RepoPresenter)},
	}

	for range iter.N(8) {
		u.wg1A.Add(1)
		go u.workerImportPath() // Phase 1 (i.e., inImportPath) to phase 2 worker.
	}
	for range iter.N(8) {
		u.wg1B.Add(1)
		go u.workerImportPathRevision() // Phase 1 (i.e., inImportPathRevision) to phase 2 worker.
	}
	go func() {
		u.wg1A.Wait()
		u.wg1B.Wait()
		close(u.phase2)
	}()

	for range iter.N(8) {
		u.wg2.Add(1)
		go u.phase23Worker() // Phase 2 to phase 3 worker.
	}
	go func() {
		u.wg2.Wait()
		close(u.phase3)
	}()

	for range iter.N(8) {
		u.wg3.Add(1)
		go u.phase34Worker() // Phase 3 to phase 4 worker.
	}
	go func() {
		u.wg3.Wait()
		close(u.out)
	}()

	go u.run()

	return u
}

// Add adds a package with specified import path for processing.
func (u *goWorkspace) Add(importPath string) {
	u.inImportPath <- importPath
}

// AddRevision adds a package with specified import path and revision for processing.
func (u *goWorkspace) AddRevision(importPath string, revision string) {
	u.inImportPathRevision <- importPathRevision{
		importPath: importPath,
		revision:   revision,
	}
}

// Done should be called after the workspace is finished being populated.
func (u *goWorkspace) Done() {
	close(u.inImportPath)
	close(u.inImportPathRevision)
}

func (u *goWorkspace) Out() <-chan *pkgs.RepoPresenter {
	response := make(chan chan *pkgs.RepoPresenter)
	u.newObserver <- observerRequest{Response: response}
	return <-response
}

func (u *goWorkspace) run() {
Outer:
	for {
		select {
		// New repoPresenter available.
		case repoPresenter, ok := <-u.out:
			// We're done streaming.
			if !ok {
				break Outer
			}

			// Append repoPresenter to current list.
			u.GoPackageList.Lock()
			u.GoPackageList.List[repoPresenter.Repo.Root] = repoPresenter
			u.GoPackageList.Unlock()

			// Send new repoPresenter to all existing observers.
			for ch := range u.observers {
				ch <- repoPresenter
			}
		// New observer request.
		case req := <-u.newObserver:
			u.GoPackageList.Lock()
			ch := make(chan *pkgs.RepoPresenter, len(u.GoPackageList.List))
			for _, repoPresenter := range u.GoPackageList.List {
				ch <- repoPresenter
			}
			u.GoPackageList.Unlock()

			u.observers[ch] = struct{}{}

			req.Response <- ch
		}
	}

	// At this point, streaming has finished, so finish up existing observers.
	for ch := range u.observers {
		close(ch)
	}
	u.observers = nil

	// Respond to new observer requests directly.
	for req := range u.newObserver {
		u.GoPackageList.Lock()
		ch := make(chan *pkgs.RepoPresenter, len(u.GoPackageList.List))
		// TODO: By now, all packages are known, so consider sorting them.
		for _, repoPresenter := range u.GoPackageList.List {
			ch <- repoPresenter
		}
		u.GoPackageList.Unlock()

		close(ch)

		req.Response <- ch
	}
}

// worker for phase 1, sends unique repos to phase 2.
func (u *goWorkspace) workerImportPath() {
	defer u.wg1A.Done()
	for importPath := range u.inImportPath {
		//started := time.Now()
		// Determine repo root and local revision.
		// This is potentially somewhat slow.
		bpkg, err := build.Import(importPath, "", build.FindOnly)
		if err != nil {
			log.Println("build.Import:", err)
			continue
		}
		//goon.DumpExpr(bpkg)
		if bpkg.Goroot {
			continue
		}
		vcs2 := vcs2.New(bpkg.Dir)
		if vcs2 == nil {
			log.Println("not in VCS:", bpkg.Dir)
			continue
		}
		repoRoot := vcs2.RootPath()[len(bpkg.SrcRoot)+1:] // TODO: Consider sym links, etc.
		//fmt.Printf("build + vcs: %v ms.\n", time.Since(started).Seconds()*1000)

		var repo *pkg.Repo
		u.reposMu.Lock()
		if _, ok := u.repos[repoRoot]; !ok {
			repo = &pkg.Repo{
				Root: repoRoot,
				Cmd:  vcs.ByCmd(vcs2.Type().VcsType()),
				VCS:  vcs2,
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

// worker for phase 1, sends unique repos to phase 2.
func (u *goWorkspace) workerImportPathRevision() {
	defer u.wg1B.Done()
	for p := range u.inImportPathRevision {
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
				Cmd:       rr.VCS,
				Local: pkg.Local{
					Revision: p.revision,
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

// Phase 2 to 3 figures out repo remote revision (and local if needed).
func (u *goWorkspace) phase23Worker() {
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
		case p.Cmd != nil: // TODO: Make this better.
			switch p.Cmd.Cmd {
			case vcs2.Git.VcsType():
				remoteVCS = vcs2.NewRemote(vcs2.Git, template.URL(p.RemoteURL))
			}
		}
		var remoteRevision string
		if remoteVCS != nil {
			remoteRevision = remoteVCS.GetRemoteRev()
		}
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

			// TODO: Organize.
			if remoteVCS != nil {
				p.Remote.IsContained = localVCS.IsContained(remoteRevision)
			}
		}

		u.phase3 <- p
	}
}

// Phase 3 to 4 worker figures out if a repo should be presented and gives it a presenter.
func (u *goWorkspace) phase34Worker() {
	defer u.wg3.Done()
	for repo := range u.phase3 {
		if !shouldPresentUpdate(repo) {
			continue
		}

		started := time.Now()

		// This part might take a while.
		repoPresenter := presenter.New(repo)

		fmt.Printf("Part 2b: %v ms.\n", time.Since(started).Seconds()*1000)

		u.out <- &pkgs.RepoPresenter{
			Repo:      repo,
			Presenter: repoPresenter,
		}
	}
}
