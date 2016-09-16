package main

import (
	"go/build"
	"log"
	"sync"

	"github.com/bradfitz/iter"
	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/presenter"
	"github.com/shurcooL/vcsstate"
	"golang.org/x/tools/go/vcs"
)

type GoPackageList struct {
	// TODO: Merge the List and OrderedList into a single struct to better communicate that it's a single data structure.
	sync.Mutex
	OrderedList []*RepoPresenter          // OrderedList has the same contents as List, but gives it a stable order.
	List        map[string]*RepoPresenter // Map key is repoRoot.
}

type RepoPresenter struct {
	Repo *pkg.Repo
	presenter.Presenter

	// TODO: Next up, use updateState with 3 states (notUpdated, updating, updated).
	//       Do that to track the intermediate state when a package is in the process
	//       of being updated.
	Updated bool
}

// workspace is a workspace environment, meaning each repo has local and remote components.
type workspace struct {
	importPaths         chan string
	importPathRevisions chan importPathRevision
	repositories        chan localRepo
	subrepos            chan subrepo

	// unique is the output of finding unique repositories from diverse possible inputs.
	unique chan *pkg.Repo
	// processedFiltered is the output of processed repos (complete with local and remote revisions),
	// with just enough information to decide if an update should be displayed.
	processedFiltered chan *pkg.Repo
	// presented is the output of processed and presented repos (complete with repo.Presenter).
	presented chan *RepoPresenter

	reposMu sync.Mutex
	repos   map[string]*pkg.Repo // Map key is the import path corresponding to the root of the repository.

	newObserver   chan observerRequest
	observers     map[chan *RepoPresenter]struct{}
	GoPackageList *GoPackageList
}

type observerRequest struct {
	Response chan chan *RepoPresenter
}

func NewWorkspace() *workspace {
	w := &workspace{
		importPaths:         make(chan string, 64),
		importPathRevisions: make(chan importPathRevision, 64),
		repositories:        make(chan localRepo, 64),
		subrepos:            make(chan subrepo, 64),
		unique:              make(chan *pkg.Repo, 64),
		processedFiltered:   make(chan *pkg.Repo, 64),
		presented:           make(chan *RepoPresenter, 64),

		repos: make(map[string]*pkg.Repo),

		newObserver:   make(chan observerRequest),
		observers:     make(map[chan *RepoPresenter]struct{}),
		GoPackageList: &GoPackageList{List: make(map[string]*RepoPresenter)},
	}

	// It is a lot of work to
	// find all Go packages in one's GOPATH workspace (or vendor.json file),
	// then group them by VCS repository,
	// and determine their local state (current revision, etc.),
	// then determine their remote state (latest remote revision, etc.),
	// then hit an API like GitHub or Gitiles to fetch descriptions of all commits
	// between the current local revision and latest remote revision for display purposes.
	//
	// That work is heavily blocked on local disk IO and network IO,
	// and also consists of dependencies. E.g., we can't ask for commit descriptions
	// until we know both the local and remote revisions, and we can't figure out local
	// revisions before we know which repository a Go package belongs to.
	//
	// Luckily, Go is great at concurrency, ʕ◔ϖ◔ʔ
	//          which also makes parallelism easy!
	// (See https://blog.golang.org/concurrency-is-not-parallelism.)
	//
	// Let's make gophers do all this work for us in multiple interconnected stages,
	// and parallelize each stage with many worker goroutines.

	// Stage 1, grouping all inputs into a set of unique repositories.
	//
	// We populate the workspace from any of the 3 sources:
	//
	// 	- via AddImportPath - import paths of Go packages from the GOPATH workspace.
	// 	- via AddRevision   - import paths of Go packages and their revisions from vendor.json or Godeps.json.
	// 	- via AddRepository - by directly adding local VCS repositories.
	// 	- via AddSubrepo    - by directly adding remote subrepos.
	//
	// The goal of processing in stage 1 is to take in diverse possible inputs
	// and convert them into a unique set of repositories for further processing by next stages.
	// When finished, all unique repositories are sent to w.unique channel
	// and the channel is closed.
	{
		var wg0 sync.WaitGroup
		for range iter.N(8) {
			wg0.Add(1)
			go w.importPathWorker(&wg0)
		}
		var wg1 sync.WaitGroup
		for range iter.N(8) {
			wg1.Add(1)
			go w.importPathRevisionWorker(&wg1)
		}
		var wg2 sync.WaitGroup
		for range iter.N(8) {
			wg2.Add(1)
			go w.repositoriesWorker(&wg2)
		}
		var wg3 sync.WaitGroup
		for range iter.N(8) {
			wg3.Add(1)
			go w.subreposWorker(&wg3)
		}
		go func() {
			wg0.Wait()
			wg1.Wait()
			wg2.Wait()
			wg3.Wait()
			close(w.unique)
		}()
	}

	// Stage 2, figuring out which repositories have updates available.
	//
	// We compute repository remote revision (and local if needed)
	// in order to figure out if repositories should be presented,
	// or filtered out (for example, because there are no updates available).
	// When finished, all non-filtered-out repositories are sent to w.processedFiltered channel
	// and the channel is closed.
	{
		var wg sync.WaitGroup
		for range iter.N(8) {
			wg.Add(1)
			go w.processFilterWorker(&wg)
		}
		go func() {
			wg.Wait()
			close(w.processedFiltered)
		}()
	}

	// Stage 3, filling in the update presentation information.
	//
	// We talk to remote APIs to fill in the missing presentation details
	// that are not available from VCS (unless we fetch commits, but we choose not to that).
	// Primarily, we get the commit messages for all the new commits that are available.
	// When finished, all repositories complete with full presentation information
	// are sent to w.presented channel and the channel is closed.
	{
		var wg sync.WaitGroup
		for range iter.N(8) {
			wg.Add(1)
			go w.presenterWorker(&wg)
		}
		go func() {
			wg.Wait()
			close(w.presented)
		}()
	}

	go w.run()

	return w
}

// AddImportPath adds a package with specified import path for processing.
func (w *workspace) AddImportPath(importPath string) {
	w.importPaths <- importPath
}

type importPathRevision struct {
	importPath string
	revision   string
}

// AddRevision adds a package with specified import path and revision for processing.
func (w *workspace) AddRevision(importPath string, revision string) {
	w.importPathRevisions <- importPathRevision{
		importPath: importPath,
		revision:   revision,
	}
}

type localRepo struct {
	Path string
	Root string
	VCS  *vcs.Cmd
}

// AddRepository adds the specified repository for processing.
func (w *workspace) AddRepository(r localRepo) {
	w.repositories <- r
}

// subrepo represents a "virtual" sub-repository inside a larger actual VCS repository.
type subrepo struct {
	Root      string
	RemoteVCS vcsstate.RemoteVCS // RemoteVCS allows getting the remote state of the VCS.
	RemoteURL string             // RemoteURL is the remote URL, including scheme.
	Revision  string
}

// AddSubrepo adds the specified subrepo for processing.
func (w *workspace) AddSubrepo(s subrepo) {
	w.subrepos <- s
}

// Done should be called after the workspace is finished being populated.
func (w *workspace) Done() {
	close(w.importPaths)
	close(w.importPathRevisions)
	close(w.repositories)
	close(w.subrepos)
}

// RepoPresenters returns a channel of all repo presenters.
// Repo presenters that are ready will be sent immediately.
// The remaining repo presenters will be sent onto the channel
// as they become available. Once all repo presenters have been
// sent, the channel will be closed. Therefore, iterating over
// the channel may block until all processing is done, but it
// will effectively return all repo presenters as soon as possible.
//
// It's safe to call RepoPresenters at any time and concurrently
// to get multiple such channels.
func (w *workspace) RepoPresenters() <-chan *RepoPresenter {
	response := make(chan chan *RepoPresenter)
	w.newObserver <- observerRequest{Response: response}
	return <-response
}

func (w *workspace) run() {
Outer:
	for {
		select {
		// New repoPresenter available.
		case repoPresenter, ok := <-w.presented:
			// We're done streaming.
			if !ok {
				break Outer
			}

			// Append repoPresenter to current list.
			w.GoPackageList.Lock()
			w.GoPackageList.OrderedList = append(w.GoPackageList.OrderedList, repoPresenter)
			w.GoPackageList.List[repoPresenter.Repo.Root] = repoPresenter
			w.GoPackageList.Unlock()

			// Send new repoPresenter to all existing observers.
			for ch := range w.observers {
				ch <- repoPresenter
			}
		// New observer request.
		case req := <-w.newObserver:
			w.GoPackageList.Lock()
			ch := make(chan *RepoPresenter, len(w.GoPackageList.OrderedList))
			for _, repoPresenter := range w.GoPackageList.OrderedList {
				ch <- repoPresenter
			}
			w.GoPackageList.Unlock()

			w.observers[ch] = struct{}{}

			req.Response <- ch
		}
	}

	// At this point, streaming has finished, so finish up existing observers.
	for ch := range w.observers {
		close(ch)
	}
	w.observers = nil

	// Respond to new observer requests directly.
	for req := range w.newObserver {
		w.GoPackageList.Lock()
		ch := make(chan *RepoPresenter, len(w.GoPackageList.OrderedList))
		for _, repoPresenter := range w.GoPackageList.OrderedList {
			ch <- repoPresenter
		}
		w.GoPackageList.Unlock()

		close(ch)

		req.Response <- ch
	}
}

// importPathWorker sends unique repositories to phase 2.
func (w *workspace) importPathWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for importPath := range w.importPaths {
		// Determine repo root.
		// This is potentially somewhat slow.
		bpkg, err := build.Import(importPath, wd, build.FindOnly|build.IgnoreVendor) // THINK: This (build.FindOnly) may find repos even when importPath has no actual package... Is that okay?
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
		w.reposMu.Lock()
		if _, ok := w.repos[root]; !ok {
			repo = &pkg.Repo{
				Root: root,

				// This is a local repository inside GOPATH. Set all of its fields.
				VCS:  vcs,
				Path: bpkg.Dir,
				Cmd:  vcsCmd,

				// TODO: Maybe keep track of import paths inside, etc.
			}
			w.repos[root] = repo
		} else {
			// TODO: Maybe keep track of import paths inside, etc.
		}
		w.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			w.unique <- repo
		}
	}
}

// importPathRevisionWorker sends unique repositories to phase 2.
func (w *workspace) importPathRevisionWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range w.importPathRevisions {
		// Determine repo root.
		// This is potentially somewhat slow.
		rr, err := vcs.RepoRootForImportPath(p.importPath, false)
		if err != nil {
			log.Printf("failed to dynamically determine repo root for %v: %v\n", p.importPath, err)
			continue
		}
		remoteVCS, err := vcsstate.NewRemoteVCS(rr.VCS)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v\n", rr.Root, err)
			continue
		}

		var repo *pkg.Repo
		w.reposMu.Lock()
		if _, ok := w.repos[rr.Root]; !ok {
			repo = &pkg.Repo{
				Root: rr.Root,

				// This is a remote repository only. Set all of its fields.
				RemoteVCS: remoteVCS,
				RemoteURL: rr.Repo,
			}
			repo.Local.Revision = p.revision
			repo.Remote.RepoURL = rr.Repo
			w.repos[rr.Root] = repo
		}
		w.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			w.unique <- repo
		}
	}
}

// repositoriesWorker sends unique repositories to phase 2.
func (w *workspace) repositoriesWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range w.repositories {
		vcsCmd, root := r.VCS, r.Root
		vcs, err := vcsstate.NewVCS(vcsCmd)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v", root, err)
			continue
		}

		var repo *pkg.Repo
		w.reposMu.Lock()
		if _, ok := w.repos[root]; !ok {
			repo = &pkg.Repo{
				Root: root,

				// This is a local repository inside GOPATH. Set all of its fields.
				VCS:  vcs,
				Path: r.Path,
				Cmd:  vcsCmd,
			}
			w.repos[root] = repo
		}
		w.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			w.unique <- repo
		}
	}
}

// subreposWorker sends unique subrepos to phase 2.
func (w *workspace) subreposWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range w.subrepos {
		var repo *pkg.Repo
		w.reposMu.Lock()
		if _, ok := w.repos[r.Root]; !ok {
			repo = &pkg.Repo{
				Root: r.Root,

				// This is a remote repository only. Set all of its fields.
				RemoteVCS: r.RemoteVCS,
				RemoteURL: r.RemoteURL,
			}
			repo.Local.RemoteURL = r.RemoteURL
			repo.Local.Revision = r.Revision
			w.repos[r.Root] = repo
		}
		w.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			w.unique <- repo
		}
	}
}

// processFilterWorker computes repository remote revision (and local if needed)
// in order to figure out if repositories should be presented.
func (w *workspace) processFilterWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for p := range w.unique {
		// Determine remote revision.
		// This is slow because it requires a network operation.
		switch {
		case p.VCS != nil:
			var err error
			p.Remote.Branch, p.Remote.Revision, err = p.VCS.RemoteBranchAndRevision(p.Path)
			if err != nil {
				log.Printf("skipping %q because of remote error:\n%v\n", p.Root, err)
				continue
			}

			if p.Local.Revision == "" {
				if r, err := p.VCS.LocalRevision(p.Path, p.Remote.Branch); err == nil {
					p.Local.Revision = r
				}
			}
			if r, err := p.VCS.RemoteURL(p.Path); err == nil {
				p.Local.RemoteURL = r
			}
			if rr, err := vcs.RepoRootForImportPath(p.Root, false); err == nil {
				p.Remote.RepoURL = rr.Repo
			}
		case p.RemoteVCS != nil:
			var err error
			p.Remote.Branch, p.Remote.Revision, err = p.RemoteVCS.RemoteBranchAndRevision(p.RemoteURL)
			if err != nil {
				log.Printf("skipping %q because of remote error:\n%v\n", p.Root, err)
				continue
			}
		default:
			panic("internal error: precondition failed, expected one of p.VCS or p.RemoteVCS to not be nil")
		}

		if !shouldPresentUpdate(p) {
			continue
		}

		w.processedFiltered <- p
	}
}

// presenterWorker works with repos that should be displayed, creating a presenter each.
func (w *workspace) presenterWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for repo := range w.processedFiltered {
		// This part might take a while.
		repoPresenter := presenter.New(repo)

		w.presented <- &RepoPresenter{
			Repo:      repo,
			Presenter: repoPresenter,
		}
	}
}
