// Package workspace contains a pipeline for processing a Go workspace.
package workspace

import (
	"fmt"
	"go/build"
	"html/template"
	"log"
	"sync"

	"github.com/bradfitz/iter"
	"github.com/shurcooL/Go-Package-Store"
	"github.com/shurcooL/gostatus/status"
	"github.com/shurcooL/vcsstate"
	"golang.org/x/tools/go/vcs"
)

type GoPackageList struct {
	// TODO: Merge the List and OrderedList into a single struct to better communicate that it's a single data structure.
	sync.Mutex
	OrderedList []*RepoPresentation          // OrderedList has the same contents as List, but gives it a stable order.
	List        map[string]*RepoPresentation // Map key is repoRoot.
}

type RepoPresentation struct {
	Repo         *gps.Repo
	Presentation *gps.Presentation

	// TODO: Next up, use updateState with 3 states (notUpdated, updating, updated).
	//       Do that to track the intermediate state when a package is in the process
	//       of being updated.
	Updated bool
}

// Pipeline for processing a Go workspace, where each repo has local and remote components.
type Pipeline struct {
	wd string // Working directory. Used to resolve relative import paths.

	// presenters are presenters registered with RegisterPresenter.
	presenters []gps.Presenter

	importPaths         chan string
	importPathRevisions chan importPathRevision
	repositories        chan LocalRepo
	subrepos            chan Subrepo

	// unique is the output of finding unique repositories from diverse possible inputs.
	unique chan *gps.Repo
	// processedFiltered is the output of processed repos (complete with local and remote revisions),
	// with just enough information to decide if an update should be displayed.
	processedFiltered chan *gps.Repo
	// presented is the output of processed and presented repos (complete with gps.Presentation).
	presented chan *RepoPresentation

	reposMu sync.Mutex
	repos   map[string]*gps.Repo // Map key is the import path corresponding to the root of the repository.

	newObserver   chan observerRequest
	observers     map[chan *RepoPresentation]struct{}
	GoPackageList *GoPackageList
}

type observerRequest struct {
	Response chan chan *RepoPresentation
}

// NewPipeline creates a Pipeline with working directory wd.
// Working directory is used to resolve relative import paths.
//
// First, available presenters should be registered via RegisterPresenter.
// Then Go packages can be added via various means. Call Done once done adding.
// Processing begins as soon as Go packages are added to the pipeline.
// Results can be accessed via RepoPresentations at any time, as often as needed.
func NewPipeline(wd string) *Pipeline {
	p := &Pipeline{
		wd: wd,

		importPaths:         make(chan string, 64),
		importPathRevisions: make(chan importPathRevision, 64),
		repositories:        make(chan LocalRepo, 64),
		subrepos:            make(chan Subrepo, 64),
		unique:              make(chan *gps.Repo, 64),
		processedFiltered:   make(chan *gps.Repo, 64),
		presented:           make(chan *RepoPresentation, 64),

		repos: make(map[string]*gps.Repo),

		newObserver:   make(chan observerRequest),
		observers:     make(map[chan *RepoPresentation]struct{}),
		GoPackageList: &GoPackageList{List: make(map[string]*RepoPresentation)},
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
	// When finished, all unique repositories are sent to p.unique channel
	// and the channel is closed.
	{
		var wg0 sync.WaitGroup
		for range iter.N(8) {
			wg0.Add(1)
			go p.importPathWorker(&wg0)
		}
		var wg1 sync.WaitGroup
		for range iter.N(8) {
			wg1.Add(1)
			go p.importPathRevisionWorker(&wg1)
		}
		var wg2 sync.WaitGroup
		for range iter.N(8) {
			wg2.Add(1)
			go p.repositoriesWorker(&wg2)
		}
		var wg3 sync.WaitGroup
		for range iter.N(8) {
			wg3.Add(1)
			go p.subreposWorker(&wg3)
		}
		go func() {
			wg0.Wait()
			wg1.Wait()
			wg2.Wait()
			wg3.Wait()
			close(p.unique)
		}()
	}

	// Stage 2, figuring out which repositories have updates available.
	//
	// We compute repository remote revision (and local if needed)
	// in order to figure out if repositories should be presented,
	// or filtered out (for example, because there are no updates available).
	// When finished, all non-filtered-out repositories are sent to p.processedFiltered channel
	// and the channel is closed.
	{
		var wg sync.WaitGroup
		for range iter.N(8) {
			wg.Add(1)
			go p.processFilterWorker(&wg)
		}
		go func() {
			wg.Wait()
			close(p.processedFiltered)
		}()
	}

	// Stage 3, filling in the update presentation information.
	//
	// We talk to remote APIs to fill in the missing presentation details
	// that are not available from VCS (unless we fetch commits, but we choose not to that).
	// Primarily, we get the commit messages for all the new commits that are available.
	// When finished, all repositories complete with full presentation information
	// are sent to p.presented channel and the channel is closed.
	{
		var wg sync.WaitGroup
		for range iter.N(8) {
			wg.Add(1)
			go p.presentWorker(&wg)
		}
		go func() {
			wg.Wait()
			close(p.presented)
		}()
	}

	go p.run()

	return p
}

// RegisterPresenter registers a presenter.
// Presenters are consulted in the same order that they were registered.
func (p *Pipeline) RegisterPresenter(pr gps.Presenter) {
	p.presenters = append(p.presenters, pr)
}

// AddImportPath adds a package with specified import path for processing.
func (p *Pipeline) AddImportPath(importPath string) {
	p.importPaths <- importPath
}

type importPathRevision struct {
	importPath string
	revision   string
}

// AddRevision adds a package with specified import path and revision for processing.
func (p *Pipeline) AddRevision(importPath string, revision string) {
	p.importPathRevisions <- importPathRevision{
		importPath: importPath,
		revision:   revision,
	}
}

type LocalRepo struct {
	Path string
	Root string
	VCS  *vcs.Cmd
}

// AddRepository adds the specified repository for processing.
func (p *Pipeline) AddRepository(r LocalRepo) {
	p.repositories <- r
}

// Subrepo represents a "virtual" sub-repository inside a larger actual VCS repository.
type Subrepo struct {
	Root      string
	RemoteVCS vcsstate.RemoteVCS // RemoteVCS allows getting the remote state of the VCS.
	RemoteURL string             // RemoteURL is the remote URL, including scheme.
	Revision  string
}

// AddSubrepo adds the specified Subrepo for processing.
func (p *Pipeline) AddSubrepo(s Subrepo) {
	p.subrepos <- s
}

// Done should be called after the workspace is finished being populated.
func (p *Pipeline) Done() {
	close(p.importPaths)
	close(p.importPathRevisions)
	close(p.repositories)
	close(p.subrepos)
}

// AddPresented adds a RepoPresentation the pipeline.
// It enables mocks to directly add presented repos.
func (p *Pipeline) AddPresented(r *RepoPresentation) {
	p.presented <- r
}

// RepoPresentations returns a channel of all repo presentations.
// Repo presentations that are ready will be sent immediately.
// The remaining repo presentations will be sent onto the channel
// as they become available. Once all repo presentations have been
// sent, the channel will be closed. Therefore, iterating over
// the channel may block until all processing is done, but it
// will effectively return all repo presentations as soon as possible.
//
// It's safe to call RepoPresentations at any time and concurrently
// to get multiple such channels.
func (p *Pipeline) RepoPresentations() <-chan *RepoPresentation {
	response := make(chan chan *RepoPresentation)
	p.newObserver <- observerRequest{Response: response}
	return <-response
}

func (p *Pipeline) run() {
Outer:
	for {
		select {
		// New repoPresentation available.
		case repoPresentation, ok := <-p.presented:
			// We're done streaming.
			if !ok {
				break Outer
			}

			// Append repoPresentation to current list.
			p.GoPackageList.Lock()
			p.GoPackageList.OrderedList = append(p.GoPackageList.OrderedList, repoPresentation)
			p.GoPackageList.List[repoPresentation.Repo.Root] = repoPresentation
			p.GoPackageList.Unlock()

			// Send new repoPresentation to all existing observers.
			for ch := range p.observers {
				ch <- repoPresentation
			}
		// New observer request.
		case req := <-p.newObserver:
			p.GoPackageList.Lock()
			ch := make(chan *RepoPresentation, len(p.GoPackageList.OrderedList))
			for _, repoPresentation := range p.GoPackageList.OrderedList {
				ch <- repoPresentation
			}
			p.GoPackageList.Unlock()

			p.observers[ch] = struct{}{}

			req.Response <- ch
		}
	}

	// At this point, streaming has finished, so finish up existing observers.
	for ch := range p.observers {
		close(ch)
	}
	p.observers = nil

	// Respond to new observer requests directly.
	for req := range p.newObserver {
		p.GoPackageList.Lock()
		ch := make(chan *RepoPresentation, len(p.GoPackageList.OrderedList))
		for _, repoPresentation := range p.GoPackageList.OrderedList {
			ch <- repoPresentation
		}
		p.GoPackageList.Unlock()

		close(ch)

		req.Response <- ch
	}
}

// importPathWorker sends unique repositories to phase 2.
func (p *Pipeline) importPathWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for importPath := range p.importPaths {
		// Determine repo root.
		// This is potentially somewhat slow.
		bpkg, err := build.Import(importPath, p.wd, build.FindOnly|build.IgnoreVendor) // THINK: This (build.FindOnly) may find repos even when importPath has no actual package... Is that okay?
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

		var repo *gps.Repo
		p.reposMu.Lock()
		if _, ok := p.repos[root]; !ok {
			repo = &gps.Repo{
				Root: root,

				// This is a local repository inside GOPATH. Set all of its fields.
				VCS:  vcs,
				Path: bpkg.Dir,
				Cmd:  vcsCmd,

				// TODO: Maybe keep track of import paths inside, etc.
			}
			p.repos[root] = repo
		} else {
			// TODO: Maybe keep track of import paths inside, etc.
		}
		p.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			p.unique <- repo
		}
	}
}

// importPathRevisionWorker sends unique repositories to phase 2.
func (p *Pipeline) importPathRevisionWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for ipr := range p.importPathRevisions {
		// Determine repo root.
		// This is potentially somewhat slow.
		rr, err := vcs.RepoRootForImportPath(ipr.importPath, false)
		if err != nil {
			log.Printf("failed to dynamically determine repo root for %v: %v\n", ipr.importPath, err)
			continue
		}
		remoteVCS, err := vcsstate.NewRemoteVCS(rr.VCS)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v\n", rr.Root, err)
			continue
		}

		var repo *gps.Repo
		p.reposMu.Lock()
		if _, ok := p.repos[rr.Root]; !ok {
			repo = &gps.Repo{
				Root: rr.Root,

				// This is a remote repository only. Set all of its fields.
				RemoteVCS: remoteVCS,
				RemoteURL: rr.Repo,
			}
			repo.Local.Revision = ipr.revision
			repo.Remote.RepoURL = rr.Repo
			p.repos[rr.Root] = repo
		}
		p.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			p.unique <- repo
		}
	}
}

// repositoriesWorker sends unique repositories to phase 2.
func (p *Pipeline) repositoriesWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range p.repositories {
		vcsCmd, root := r.VCS, r.Root
		vcs, err := vcsstate.NewVCS(vcsCmd)
		if err != nil {
			log.Printf("repo %v not supported by vcsstate: %v", root, err)
			continue
		}

		var repo *gps.Repo
		p.reposMu.Lock()
		if _, ok := p.repos[root]; !ok {
			repo = &gps.Repo{
				Root: root,

				// This is a local repository inside GOPATH. Set all of its fields.
				VCS:  vcs,
				Path: r.Path,
				Cmd:  vcsCmd,
			}
			p.repos[root] = repo
		}
		p.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			p.unique <- repo
		}
	}
}

// subreposWorker sends unique subrepos to phase 2.
func (p *Pipeline) subreposWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range p.subrepos {
		// Determine repo root.
		// This is potentially somewhat slow.
		rr, err := vcs.RepoRootForImportPath(r.Root, false)
		if err != nil {
			log.Printf("failed to dynamically determine repo root for %v: %v\n", r.Root, err)
			continue
		}

		var repo *gps.Repo
		p.reposMu.Lock()
		if _, ok := p.repos[r.Root]; !ok {
			repo = &gps.Repo{
				Root: r.Root,

				// This is a remote repository only. Set all of its fields.
				RemoteVCS: r.RemoteVCS,
				RemoteURL: r.RemoteURL,
			}
			repo.Local.RemoteURL = r.RemoteURL // TODO: Consider having r.RemoteURL take precedence over rr.Repo. But need to make that play nicely with the updaters; see TODO at bottom of gps.Repo struct.
			repo.Local.Revision = r.Revision
			repo.Remote.RepoURL = rr.Repo
			p.repos[r.Root] = repo
		}
		p.reposMu.Unlock()

		// If new repo, send off to phase 2 channel.
		if repo != nil {
			p.unique <- repo
		}
	}
}

// processFilterWorker computes repository remote revision (and local if needed)
// in order to figure out if repositories should be presented.
func (p *Pipeline) processFilterWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range p.unique {
		// Determine remote revision.
		// This is slow because it requires a network operation.
		switch {
		case r.VCS != nil:
			var err error
			r.Remote.Branch, r.Remote.Revision, err = r.VCS.RemoteBranchAndRevision(r.Path)
			if err != nil {
				log.Printf("skipping %q because of remote error:\n\t%v\n", r.Root, err)
				continue
			}

			if r.Local.Revision == "" {
				if rev, err := r.VCS.LocalRevision(r.Path, r.Remote.Branch); err == nil {
					r.Local.Revision = rev
				}
			}
			if ru, err := r.VCS.RemoteURL(r.Path); err == nil {
				r.Local.RemoteURL = ru
			}
			if rr, err := vcs.RepoRootForImportPath(r.Root, false); err == nil {
				r.Remote.RepoURL = rr.Repo
			}
		case r.RemoteVCS != nil:
			var err error
			r.Remote.Branch, r.Remote.Revision, err = r.RemoteVCS.RemoteBranchAndRevision(r.RemoteURL)
			if err != nil {
				log.Printf("skipping %q because of remote error:\n\t%v\n", r.Root, err)
				continue
			}
		default:
			panic("internal error: precondition failed, expected one of r.VCS or r.RemoteVCS to not be nil")
		}

		if ok, reason := shouldPresentUpdate(r); !ok {
			if reason != "" {
				log.Printf("skipping %q because:\n\t%v\n", r.Root, reason)
			}
			continue
		}

		p.processedFiltered <- r
	}
}

// shouldPresentUpdate reports if the given goPackage should be presented as an available update.
// It checks that the Go package is on default branch, does not have a dirty working tree, and does not have the remote revision.
// It returns a non-empty reason for why an update should be skipped, or empty string if it's not interesting (e.g., repository is up to date).
func shouldPresentUpdate(repo *gps.Repo) (ok bool, reason string) {
	// Do some sanity checks.
	if repo.Remote.RepoURL == "" {
		return false, "repository URL (as determined dynamically from the import path) is empty"
	}
	if repo.Local.Revision == "" {
		return false, "local revision is empty"
	}
	if repo.Remote.Revision == "" {
		return false, "remote revision is empty"
	}

	if repo.Local.Revision == repo.Remote.Revision {
		// Already up to date. No reason provided because it's not worth mentioning.
		return false, ""
	}

	// Check repository state before presenting updates.
	switch {
	case repo.VCS != nil:
		// Local branch should match remote branch.
		localBranch, err := repo.VCS.Branch(repo.Path)
		if err != nil {
			return false, "error determining local branch:\n" + err.Error()
		}
		if localBranch != repo.Remote.Branch {
			return false, fmt.Sprintf("local branch %q doesn't match remote branch %q", localBranch, repo.Remote.Branch)
		}

		// There shouldn't be a dirty working tree.
		treeStatus, err := repo.VCS.Status(repo.Path)
		if err != nil {
			return false, "error determining if working tree is dirty:\n" + err.Error()
		}
		if treeStatus != "" {
			return false, "working tree is dirty:\n" + treeStatus
		}

		// Local remote URL should match Repo URL derived from import path.
		if !status.EqualRepoURLs(repo.Local.RemoteURL, repo.Remote.RepoURL) {
			return false, fmt.Sprintf("remote URL (%s) doesn't match repo URL inferred from import path (%s)", repo.Local.RemoteURL, repo.Remote.RepoURL)
		}

		// The local commit should be contained by remote. Otherwise, it means the local
		// repository commit is actually ahead of remote, and there's nothing to update (instead, the
		// user probably needs to push their local work to remote).
		localContainsRemoteRevision, err := repo.VCS.Contains(repo.Path, repo.Remote.Revision, repo.Remote.Branch)
		if err != nil {
			return false, "error determining if local commit is contained by remote:\n" + err.Error()
		}
		if localContainsRemoteRevision {
			return false, fmt.Sprintf("local revision %q is ahead of remote revision %q", repo.Local.Revision, repo.Remote.Revision)
		}

	case repo.RemoteVCS != nil:
		// TODO: Consider taking care of this difference in remote URLs earlier, inside, e.g., subreposWorker. But need to make that play nicely with the updaters; see TODO at bottom of gps.Repo struct.
		//
		// Local remote URL, if set, should match Repo URL derived from import path.
		if repo.Local.RemoteURL != "" && !status.EqualRepoURLs(repo.Local.RemoteURL, repo.Remote.RepoURL) {
			return false, fmt.Sprintf("remote URL (%s) doesn't match repo URL inferred from import path (%s)", repo.Local.RemoteURL, repo.Remote.RepoURL)
		}
	}

	// If we got this far, there's an update available and everything looks normal. Present it.
	return true, ""
}

// presentWorker works with repos that should be displayed, creating a presentation for each.
func (p *Pipeline) presentWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for repo := range p.processedFiltered {
		// This part might take a while.
		presentation := p.present(repo)

		p.presented <- &RepoPresentation{
			Repo:         repo,
			Presentation: presentation,
		}
	}
}

// present takes a repository containing 1 or more Go packages, and returns a presentation for it.
// It tries to find the best presenter for the given repository out of the registered ones,
// but falls back to a generic presentation if there's nothing better.
func (p *Pipeline) present(repo *gps.Repo) *gps.Presentation {
	for _, presenter := range p.presenters {
		if presentation := presenter(repo); presentation != nil {
			return presentation
		}
	}

	// Generic presentation.
	return &gps.Presentation{
		Home:    template.URL("https://" + repo.Root),
		Image:   "https://github.com/images/gravatars/gravatar-user-420.png",
		Changes: nil,
		Error:   nil,
	}
}
