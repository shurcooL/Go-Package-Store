package main

import (
	"fmt"
	"go/build"
	"log"
	"sync"
	"time"

	"github.com/bradfitz/iter"
	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/pkgs"
	"github.com/shurcooL/Go-Package-Store/presenter"
	"github.com/shurcooL/vcsstate"
	"golang.org/x/tools/go/vcs"
)

type importPathRevision struct {
	importPath string
	revision   string
}

// workspace is a workspace environment, meaning each repo has local and remote components.
type workspace struct {
	inImportPath         chan string
	inRepos              chan Repo
	inImportPathRevision chan importPathRevision
	phase2               chan *pkg.Repo
	// phase3 is the output of processed repos (complete with local and remote revisions),
	// with just enough information to decide if an update should be displayed.
	phase3 chan *pkg.Repo
	// out is the output of processed and presented repos (complete with repo.Presenter).
	out chan *pkgs.RepoPresenter

	reposMu sync.Mutex
	repos   map[string]*pkg.Repo // Map key is the import path corresponding to the root of the repository or Go package.

	newObserver   chan observerRequest
	observers     map[chan *pkgs.RepoPresenter]struct{}
	GoPackageList *pkgs.GoPackageList
}

type observerRequest struct {
	Response chan chan *pkgs.RepoPresenter
}

func NewWorkspace() *workspace {
	w := &workspace{
		inImportPath:         make(chan string, 64),
		inRepos:              make(chan Repo, 64),
		inImportPathRevision: make(chan importPathRevision, 64),
		phase2:               make(chan *pkg.Repo, 64),
		phase3:               make(chan *pkg.Repo, 64),
		out:                  make(chan *pkgs.RepoPresenter, 64),

		repos: make(map[string]*pkg.Repo),

		newObserver:   make(chan observerRequest),
		observers:     make(map[chan *pkgs.RepoPresenter]struct{}),
		GoPackageList: &pkgs.GoPackageList{List: make(map[string]*pkgs.RepoPresenter)},
	}

	{
		var wg0 sync.WaitGroup
		for range iter.N(8) {
			wg0.Add(1)
			go w.workerImportPath(&wg0)
		}
		var wg1 sync.WaitGroup
		for range iter.N(8) {
			wg1.Add(1)
			go w.workerRepos(&wg1)
		}
		var wg2 sync.WaitGroup
		for range iter.N(8) {
			wg2.Add(1)
			go w.workerImportPathRevision(&wg2)
		}
		go func() {
			wg0.Wait()
			wg1.Wait()
			wg2.Wait()
			close(w.phase2)
			fmt.Println("time.Since(startedPhase1):", time.Since(startedPhase1))
			fmt.Println("alreadyEnteredPkgs:", alreadyEnteredPkgs)
		}()
	}

	//return w

	{
		var wg sync.WaitGroup
		for range iter.N(8) {
			wg.Add(1)
			go w.phase23Worker(&wg)
		}
		go func() {
			wg.Wait()
			close(w.phase3)
		}()
	}

	{
		var wg sync.WaitGroup
		for range iter.N(8) {
			wg.Add(1)
			go w.phase34Worker(&wg)
		}
		go func() {
			wg.Wait()
			close(w.out)
		}()
	}

	go w.run()

	return w
}

// Add adds a package with specified import path for processing.
func (u *workspace) Add(importPath string) {
	u.inImportPath <- importPath
}

func (u *workspace) AddRepo(r Repo) {
	u.inRepos <- r
}

// AddRevision adds a package with specified import path and revision for processing.
func (u *workspace) AddRevision(importPath string, revision string) {
	u.inImportPathRevision <- importPathRevision{
		importPath: importPath,
		revision:   revision,
	}
}

// Done should be called after the workspace is finished being populated.
func (u *workspace) Done() {
	close(u.inImportPath)
	close(u.inRepos)
	close(u.inImportPathRevision)
}

func (u *workspace) Out() <-chan *pkgs.RepoPresenter {
	response := make(chan chan *pkgs.RepoPresenter)
	u.newObserver <- observerRequest{Response: response}
	return <-response
}

func (u *workspace) run() {
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
			u.GoPackageList.OrderedList = append(u.GoPackageList.OrderedList, repoPresenter)
			u.GoPackageList.List[repoPresenter.Repo.Root] = repoPresenter
			u.GoPackageList.Unlock()

			// Send new repoPresenter to all existing observers.
			for ch := range u.observers {
				ch <- repoPresenter
			}
		// New observer request.
		case req := <-u.newObserver:
			u.GoPackageList.Lock()
			ch := make(chan *pkgs.RepoPresenter, len(u.GoPackageList.OrderedList))
			for _, repoPresenter := range u.GoPackageList.OrderedList {
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
		ch := make(chan *pkgs.RepoPresenter, len(u.GoPackageList.OrderedList))
		// TODO: By now, all packages are known, so consider sorting them.
		for _, repoPresenter := range u.GoPackageList.OrderedList {
			ch <- repoPresenter
		}
		u.GoPackageList.Unlock()

		close(ch)

		req.Response <- ch
	}
}

var alreadyEnteredPkgs = 0

type Repo struct {
	Path string
	Root string
	VCS  *vcs.Cmd
}

// worker for phase 1, sends unique repos to phase 2.
func (u *workspace) workerRepos(wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range u.inRepos {
		vcsCmd, root := r.VCS, r.Root
		vcs, err := vcsstate.NewVCS(vcsCmd)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v", root, err)
			continue
		}

		var repo *pkg.Repo
		u.reposMu.Lock()
		if _, ok := u.repos[root]; !ok {
			repo = &pkg.Repo{
				Path: r.Path,
				Root: root,
				Cmd:  vcsCmd,
				VCS:  vcs,
				// TODO: Maybe keep track of import paths inside, etc.
			}
			u.repos[root] = repo
		} else {
			// TODO: Maybe keep track of import paths inside, etc.
		}
		u.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			u.phase2 <- repo
		} else {
			alreadyEnteredPkgs++
		}
	}
}

// worker for phase 1, sends unique repos to phase 2.
func (u *workspace) workerImportPath(wg *sync.WaitGroup) {
	defer wg.Done()
	for importPath := range u.inImportPath {
		// Determine repo root.
		// This is potentially somewhat slow.
		bpkg, err := build.Import(importPath, "", build.FindOnly)
		if err != nil {
			log.Println("build.Import:", err)
			continue
		}
		if bpkg.Goroot {
			// Go-Package-Store has no support for updating packages in GOROOT, so skip those.
			continue
		}
		vcsCmd, root, err := vcs.FromDir(bpkg.Dir, bpkg.SrcRoot)
		if err != nil {
			// Go package not under VCS.
			continue
		}
		vcs, err := vcsstate.NewVCS(vcsCmd)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v", root, err)
			continue
		}

		var repo *pkg.Repo
		u.reposMu.Lock()
		if _, ok := u.repos[root]; !ok {
			repo = &pkg.Repo{
				Path: bpkg.Dir,
				Root: root,
				Cmd:  vcsCmd,
				VCS:  vcs,
				// TODO: Maybe keep track of import paths inside, etc.
			}
			u.repos[root] = repo
		} else {
			// TODO: Maybe keep track of import paths inside, etc.
		}
		u.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			u.phase2 <- repo
		} else {
			alreadyEnteredPkgs++
		}
	}
}

// worker for phase 1, sends unique repos to phase 2.
func (u *workspace) workerImportPathRevision(wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range u.inImportPathRevision {
		//started := time.Now()
		// Determine repo root.
		// This is potentially somewhat slow.
		rr, err := vcs.RepoRootForImportPath(p.importPath, false)
		if err != nil {
			panic(err) // TODO.
		}
		//fmt.Printf("rr: %v ms.\n", time.Since(started).Seconds()*1000)
		remoteVCS, err := vcsstate.NewRemoteVCS(rr.VCS)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v\n", rr.Root, err)
			continue
		}

		var repo *pkg.Repo
		u.reposMu.Lock()
		if _, ok := u.repos[rr.Root]; !ok {
			repo = &pkg.Repo{
				Root:      rr.Root,
				RemoteURL: rr.Repo,
				Cmd:       rr.VCS,
				RemoteVCS: remoteVCS,
				// TODO: Maybe keep track of import paths inside, etc.
			}
			repo.Local.Revision = p.revision
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

// Phase 2 to 3 figures out repo remote revision (and local if needed)
// in order to figure out if a repo should be presented.
func (u *workspace) phase23Worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range u.phase2 {
		//started := time.Now()
		// Determine remote revision.
		// This is slow because it requires a network operation.
		var remoteRevision string
		if p.VCS != nil {
			var err error
			remoteRevision, err = p.VCS.RemoteRevision(p.Path)
			_ = err // TODO.
		} else if p.RemoteVCS != nil {
			var err error
			remoteRevision, err = p.RemoteVCS.RemoteRevision(p.RemoteURL)
			_ = err // TODO.
		}
		//fmt.Printf("remoteVCS.GetRemoteRev: %v ms.\n", time.Since(started).Seconds()*1000)

		p.Remote.Revision = remoteRevision

		// TODO: Organize.
		if p.Local.Revision == "" && p.VCS != nil {
			if r, err := p.VCS.LocalRevision(p.Path); err == nil {
				p.Local.Revision = r
			}

			// TODO: Organize.
			if p.RemoteVCS == nil && p.RemoteURL == "" {
				if r, err := p.VCS.RemoteURL(p.Path); err == nil {
					p.RemoteURL = r
				}
			}
		}

		if !shouldPresentUpdate(p) {
			continue
		}

		u.phase3 <- p
	}
}

// Phase 3 to 4 worker works with repos that should be displayed, creating a presenter each.
func (u *workspace) phase34Worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for repo := range u.phase3 {
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
