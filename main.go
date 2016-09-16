// Go Package Store displays updates for the Go packages in your GOPATH.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/presenter/github"
	"github.com/shurcooL/Go-Package-Store/presenter/gitiles"
	"github.com/shurcooL/Go-Package-Store/repo"
	"github.com/shurcooL/go/open"
	"github.com/shurcooL/go/ospath"
	"github.com/shurcooL/gostatus/status"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"github.com/shurcooL/httpgzip"
	"golang.org/x/net/websocket"
	"golang.org/x/oauth2"

	// Register presenters.
	_ "github.com/shurcooL/Go-Package-Store/presenter/github"
	_ "github.com/shurcooL/Go-Package-Store/presenter/gitiles"
)

// shouldPresentUpdate determines if the given goPackage should be presented as an available update.
// It checks that the Go package is on default branch, does not have a dirty working tree, and does not have the remote revision.
func shouldPresentUpdate(repo *pkg.Repo) bool {
	if repo.Remote.RepoURL == "" || repo.Local.Revision == "" || repo.Remote.Revision == "" {
		return false
	}

	// If this is a local VCS repository, do some sanity checks before presenting updates.
	if repo.VCS != nil {
		// Local branch should match remote branch.
		if localBranch, err := repo.VCS.Branch(repo.Path); err != nil || localBranch != repo.Remote.Branch {
			return false
		}
		// There shouldn't be a dirty working tree.
		if status, err := repo.VCS.Status(repo.Path); err != nil || status != "" {
			return false
		}
		// Local remote URL should match Repo URL derived from import path.
		if !status.EqualRepoURLs(repo.Local.RemoteURL, repo.Remote.RepoURL) {
			return false
		}
		// The local commit should be contained by remote. Otherwise, it means the local
		// repository commit is actually ahead of remote, and there's nothing to update (instead, the
		// user probably needs to push their local work to remote).
		if c, err := repo.VCS.Contains(repo.Path, repo.Remote.Revision, repo.Remote.Branch); err != nil || c {
			return false
		}
	}

	return repo.Local.Revision != repo.Remote.Revision
}

var (
	pipeline *workspace = NewWorkspace()

	// updater is set based on the source of Go packages. If nil, it means
	// we don't have support to update Go packages from the current source.
	// It's used to update repos in the backend, and if set to nil, to disable
	// the frontend UI for updating packages.
	updater repo.Updater
)

type updateRequest struct {
	root       string
	resultChan chan error
}

var updateRequestChan = make(chan updateRequest)

// updateWorker is a sequential updater of Go packages. It does not update them in parallel
// to avoid race conditions or other problems, since `go get -u` does not seem to protect against that.
func updateWorker() {
	for updateRequest := range updateRequestChan {
		pipeline.GoPackageList.Lock()
		repoPresenter, ok := pipeline.GoPackageList.List[updateRequest.root]
		pipeline.GoPackageList.Unlock()
		if !ok {
			updateRequest.resultChan <- fmt.Errorf("root %q not found", updateRequest.root)
			continue
		}
		if repoPresenter.Updated {
			updateRequest.resultChan <- fmt.Errorf("root %q already updated", updateRequest.root)
			continue
		}

		updateResult := updater.Update(repoPresenter.Repo)
		if updateResult == nil {
			// Mark repo as updated.
			pipeline.GoPackageList.Lock()
			// Move it down the OrderedList towards all other updated.
			{
				var i, j int
				for ; pipeline.GoPackageList.OrderedList[i].Repo.Root != updateRequest.root; i++ { // i is the current package about to be updated.
				}
				for j = len(pipeline.GoPackageList.OrderedList) - 1; pipeline.GoPackageList.OrderedList[j].Updated; j-- { // j is the last not-updated package.
				}
				pipeline.GoPackageList.OrderedList[i], pipeline.GoPackageList.OrderedList[j] =
					pipeline.GoPackageList.OrderedList[j], pipeline.GoPackageList.OrderedList[i]
			}
			pipeline.GoPackageList.List[updateRequest.root].Updated = true
			pipeline.GoPackageList.Unlock()
		}
		updateRequest.resultChan <- updateResult
		fmt.Println("\nDone.")
	}
}

// updateHandler is the handler for update requests.
func updateHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method should be POST.", http.StatusMethodNotAllowed)
		return
	}

	root := req.PostFormValue("repo_root")

	updateRequest := updateRequest{
		root:       root,
		resultChan: make(chan error),
	}
	updateRequestChan <- updateRequest

	err := <-updateRequest.resultChan
	// TODO: Display error in frontend.
	if err != nil {
		log.Println(err)
	}
}

// mainHandler is the handler for the index page.
func mainHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method should be GET", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	data := struct {
		Production bool
		HTTPAddr   string
	}{
		Production: production,
		HTTPAddr:   *httpFlag,
	}
	err := t.ExecuteTemplate(w, "head.html.tmpl", data)
	if err != nil {
		log.Println("ExecuteTemplate head.html.tmpl:", err)
		return
	}

	flusher := w.(http.Flusher)
	flusher.Flush()

	var updatesAvailable = 0
	var wroteInstalledUpdatesHeader bool

	for repoPresenter := range pipeline.RepoPresenters() {
		if !repoPresenter.Updated {
			updatesAvailable++
		}

		if repoPresenter.Updated && !wroteInstalledUpdatesHeader {
			// Make 'Installed Updates' header visible now.
			io.WriteString(w, `<div id="installed_updates"><h3 style="text-align: center;">Installed Updates</h3></div>`)

			wroteInstalledUpdatesHeader = true
		}

		err := t.ExecuteTemplate(w, "repo.html.tmpl", repoPresenter)
		if err != nil {
			log.Println("ExecuteTemplate repo.html.tmpl:", err)
			return
		}

		flusher.Flush()
	}

	if !wroteInstalledUpdatesHeader {
		// TODO: Make installed_updates available before all packages finish loading, so that it works when you update a package early. This will likely require a fully dynamically rendered frontend.
		// Append 'Installed Updates' header, but keep it hidden.
		io.WriteString(w, `<div id="installed_updates" style="display: none;"><h3 style="text-align: center;">Installed Updates</h3></div>`)
	}

	if updatesAvailable == 0 {
		io.WriteString(w, `<script>document.getElementById("no_updates").style.display = "";</script>`)
	}

	err = t.ExecuteTemplate(w, "tail.html.tmpl", nil)
	if err != nil {
		log.Println("ExecuteTemplate tail.html.tmpl:", err)
		return
	}
}

// WebSocket handler, to exit when client tab is closed.
func openedHandler(ws *websocket.Conn) {
	// Wait until connection is closed.
	io.Copy(ioutil.Discard, ws)

	//fmt.Println("Exiting, since the client tab was closed (detected closed WebSocket connection).")
	//close(updateRequestChan)
}

// ---

var t *template.Template

func loadTemplates() error {
	var err error
	t = template.New("").Funcs(template.FuncMap{
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
		"updateSupported": func() bool { return updater != nil },
		"commitID":        func(commitID string) string { return commitID[:8] },
	})
	t, err = vfstemplate.ParseGlob(assets, t, "/assets/*.tmpl")
	return err
}

var (
	httpFlag       = flag.String("http", "localhost:7043", "Listen for HTTP connections on this address.")
	stdinFlag      = flag.Bool("stdin", false, "Read the list of newline separated Go packages from stdin.")
	godepsFlag     = flag.String("godeps", "", "Read the list of Go packages from the specified Godeps.json file.")
	govendorFlag   = flag.String("govendor", "", "Read the list of Go packages from the specified vendor.json file.")
	gitSubrepoFlag = flag.String("git-subrepo", "", "Look for Go packages vendored using git-subrepo in the specified vendor directory.")
)

var wd = func() string {
	// Get current directory.
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("failed to get current directory:", err)
	}
	return wd
}()

func usage() {
	fmt.Fprint(os.Stderr, "Usage: Go-Package-Store [flags]\n")
	fmt.Fprint(os.Stderr, "       [newline separated packages] | Go-Package-Store -stdin [flags]\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, `
Examples:
  # Check for updates for all Go packages in GOPATH.
  Go-Package-Store

  # Show updates for all dependencies (recursive) of package in cur working dir.
  go list -f '{{join .Deps "\n"}}' . | Go-Package-Store -stdin

  # Show updates for all dependencies listed in vendor.json file.
  Go-Package-Store -govendor /path/to/vendor.json
`)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	// If we can have access to a cache directory on this system, use it for
	// caching HTTP requests of presenters.
	cacheDir, err := ospath.CacheDir("github.com/shurcooL/Go-Package-Store")
	if err != nil {
		log.Println("skipping persistent on-disk caching, because unable to acquire a cache dir:", err)
		cacheDir = ""
	}

	// Set GitHub presenter client.
	{
		var transport http.RoundTripper

		// Optionally, perform GitHub API authentication with provided token.
		if token := os.Getenv("GO_PACKAGE_STORE_GITHUB_TOKEN"); token != "" {
			transport = &oauth2.Transport{
				Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
			}
		}

		if cacheDir != "" {
			transport = &httpcache.Transport{
				Transport:           transport,
				Cache:               diskcache.New(filepath.Join(cacheDir, "github-presenter")),
				MarkCachedResponses: true,
			}
		}

		github.SetClient(&http.Client{Transport: transport})
	}

	// Set Gitiles presenter client.
	{
		var transport http.RoundTripper

		if cacheDir != "" {
			transport = &httpcache.Transport{
				Transport:           transport,
				Cache:               diskcache.New(filepath.Join(cacheDir, "gitiles-presenter")),
				MarkCachedResponses: true,
			}
		}

		gitiles.SetClient(&http.Client{Transport: transport})
	}

	switch {
	case !production:
		fmt.Println("Using no real packages (hit /mock.html endpoint for mocks).")
		pipeline.Done()
		updater = repo.MockUpdater{}
	default:
		fmt.Println("Using all Go packages in GOPATH.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			forEachRepository(func(r localRepo) {
				pipeline.AddRepository(r)
			})
			pipeline.Done()
		}()
		updater = repo.GopathUpdater{}
	case *stdinFlag:
		fmt.Println("Reading the list of newline separated Go packages from stdin.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			br := bufio.NewReader(os.Stdin)
			for line, err := br.ReadString('\n'); err == nil; line, err = br.ReadString('\n') {
				importPath := line[:len(line)-1] // Trim last newline.
				pipeline.AddImportPath(importPath)
			}
			pipeline.Done()
		}()
		updater = repo.GopathUpdater{}
	case *godepsFlag != "":
		fmt.Println("Reading the list of Go packages from Godeps.json file:", *godepsFlag)
		g, err := readGodeps(*godepsFlag)
		if err != nil {
			log.Fatalln("Failed to read Godeps.json file", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range g.Deps {
				pipeline.AddRevision(dependency.ImportPath, dependency.Rev)
			}
			pipeline.Done()
		}()
		updater = nil
	case *govendorFlag != "":
		fmt.Println("Reading the list of Go packages from vendor.json file:", *govendorFlag)
		v, err := readGovendor(*govendorFlag)
		if err != nil {
			log.Fatalln("Failed to read vendor.json file:", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range v.Package {
				pipeline.AddRevision(dependency.Path, dependency.Revision)
			}
			pipeline.Done()
		}()
		if gu, err := repo.NewGovendorUpdater(""); err == nil {
			updater = gu
		} else {
			log.Println("govendor updater is not available:", err)
		}
	case *gitSubrepoFlag != "":
		if _, err := exec.LookPath("git"); err != nil {
			log.Fatalln(fmt.Errorf("git binary is required for updating, but not available: %v", err))
		}
		fmt.Println("Using Go packages vendored using git-subrepo in the specified vendor directory.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			err := forEachGitSubrepo(*gitSubrepoFlag, func(s subrepo) {
				fmt.Printf("adding subrepo: %+v\n", s)
				pipeline.AddSubrepo(s)
			})
			if err != nil {
				fmt.Println(err)
			}
			pipeline.Done()
		}()
		updater = repo.MockUpdater{}
	}

	err = loadTemplates()
	if err != nil {
		log.Fatalln("loadTemplates:", err)
	}

	http.HandleFunc("/index.html", mainHandler)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	fileServer := httpgzip.FileServer(assets, httpgzip.FileServerOptions{ServeError: httpgzip.Detailed})
	http.Handle("/assets/", fileServer)
	http.Handle("/assets/octicons/", http.StripPrefix("/assets", fileServer))
	http.Handle("/opened", websocket.Handler(openedHandler)) // Exit server when client tab is closed.
	if updater != nil {
		http.HandleFunc("/-/update", updateHandler)
		go updateWorker()
	}

	// Start listening first.
	listener, err := net.Listen("tcp", *httpFlag)
	if err != nil {
		log.Fatalf("failed to listen on %q: %v\n", *httpFlag, err)
	}

	if production {
		// Open a browser tab and navigate to the main page.
		go open.Open("http://" + *httpFlag + "/index.html")
	}

	fmt.Println("Go Package Store server is running at http://" + *httpFlag + "/index.html.")

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
