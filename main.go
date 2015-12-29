// Go Package Store displays updates for the Go packages in your GOPATH.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/build"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/presenter"
	"github.com/shurcooL/Go-Package-Store/repo"
	"github.com/shurcooL/go/gzip_file_server"
	"github.com/shurcooL/go/u/u4"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"golang.org/x/net/websocket"
	"golang.org/x/tools/go/vcs"
)

func commonHead(w io.Writer) error {
	data := struct {
		Production bool
		HTTPAddr   string
	}{
		Production: production,
		HTTPAddr:   *httpFlag,
	}
	return t.ExecuteTemplate(w, "head.html.tmpl", data)
}
func commonTail(w io.Writer) error {
	return t.ExecuteTemplate(w, "tail.html.tmpl", nil)
}

// shouldPresentUpdate determines if the given goPackage should be presented as an available update.
// It checks that the Go package is on default branch, does not have a dirty working tree, and does not have the remote revision.
func shouldPresentUpdate(repo *pkg.Repo) bool {
	//return status.PlumbingPresenterV2(goPackage)[:3] == "  +" // Ignore stash.

	if repo.RemoteURL == "" || repo.Local.Revision == "" || repo.Remote.Revision == "" {
		return false
	}

	if repo.VCS != nil {
		if c, err := repo.VCS.Contains(repo.Path, repo.Remote.Revision); err != nil || c {
			return false
		}

		if b, err := repo.VCS.Branch(repo.Path); err != nil || b != repo.VCS.DefaultBranch() {
			return false
		}
		if s, err := repo.VCS.Status(repo.Path); err != nil || s != "" {
			return false
		}
	}

	return repo.Local.Revision != repo.Remote.Revision
}

// writeRepoHTML writes a <div> presentation for an available update.
func writeRepoHTML(w http.ResponseWriter, repoPresenter presenter.Presenter) {
	err := t.ExecuteTemplate(w, "repo.html.tmpl", repoPresenter)
	if err != nil {
		log.Println("t.ExecuteTemplate:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var (
	pipeline *workspace = NewWorkspace()

	// updater is set based on the source of Go packages. If nil, it means
	// we don't have support to update Go packages from the current source.
	// It's used to update repos in the backend, and to disable the frontend UI
	// for updating packages.
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

		updateResult := updater.Update(repoPresenter.Repo)
		if updateResult == nil {
			// Delete repo from list.
			pipeline.GoPackageList.Lock()
			// TODO: Consider marking the repo as "Updated" and display it that way, etc.
			for i := range pipeline.GoPackageList.OrderedList {
				if pipeline.GoPackageList.OrderedList[i].Repo.Root == updateRequest.root {
					pipeline.GoPackageList.OrderedList = append(pipeline.GoPackageList.OrderedList[:i], pipeline.GoPackageList.OrderedList[i+1:]...)
					break
				}
			}
			delete(pipeline.GoPackageList.List, updateRequest.root)
			pipeline.GoPackageList.Unlock()
		}
		updateRequest.resultChan <- updateResult
		fmt.Println("\nDone.")
	}
}

// Handler for update requests.
func updateHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}

	importPathPattern := req.PostFormValue("import_path_pattern") // TODO: Maybe emit root directly from frontend?
	root := importPathPattern[:len(importPathPattern)-4]

	updateRequest := updateRequest{
		root:       root,
		resultChan: make(chan error),
	}
	updateRequestChan <- updateRequest

	err := <-updateRequest.resultChan
	_ = err // TODO: Maybe display error in frontend. For now, don't do anything.
	fmt.Println("update worker:", err)
}

// Main index page handler.
func mainHandler(w http.ResponseWriter, req *http.Request) {
	if err := loadTemplates(); err != nil {
		fmt.Fprintln(w, "loadTemplates:", err)
		return
	}

	fmt.Println("mainHandler:", req.Method, req.URL.Path)

	started := time.Now()

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	_ = commonHead(w)
	defer func() { _ = commonTail(w) }()

	flusher := w.(http.Flusher)
	flusher.Flush()

	fmt.Printf("Part 1: %v ms.\n", time.Since(started).Seconds()*1000)

	updatesAvailable := 0

	for out := range pipeline.Out() {
		repoPresenter := out.Presenter

		updatesAvailable++
		writeRepoHTML(w, repoPresenter)

		flusher.Flush()
	}

	if updatesAvailable == 0 {
		io.WriteString(w, `<script>document.getElementById("no_updates").style.display = "";</script>`)
	}

	fmt.Printf("Part 3: %v ms.\n", time.Since(started).Seconds()*1000)
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
		"updateSupported": func() bool { return updater != nil },
	})
	t, err = vfstemplate.ParseGlob(assets, t, "/assets/*.tmpl")
	return err
}

var (
	httpFlag     = flag.String("http", "localhost:7043", "Listen for HTTP connections on this address.")
	stdinFlag    = flag.Bool("stdin", false, "Read the list of newline separated Go packages from stdin.")
	godepsFlag   = flag.String("godeps", "", "Read the list of Go packages from the specified Godeps.json file.")
	govendorFlag = flag.String("govendor", "", "Read the list of Go packages from the specified vendor.json file.")
)

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

var startedPhase1 time.Time

func main() {
	flag.Usage = usage
	flag.Parse()

	switch {
	default:
		fmt.Println("Using all Go packages in GOPATH.")
		/*go func() { // This needs to happen in the background because sending input will be blocked on processing.
			startedPhase1 = time.Now()
			packages := 0
			buildutil.ForEachPackage(&build.Default, func(importPath string, err error) {
				if err != nil {
					log.Println(err)
					return
				}
				pipeline.Add(importPath)
				packages++
			})
			pipeline.Done()
			fmt.Printf("%v packages.\n", packages)
		}()*/
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			startedPhase1 = time.Now()
			packages := 0
			{
				for _, workspace := range filepath.SplitList(build.Default.GOPATH) {
					srcRoot := filepath.Join(workspace, "src")
					if fi, err := os.Stat(srcRoot); err != nil || !fi.IsDir() {
						continue
					}
					_ = filepath.Walk(srcRoot, func(path string, fi os.FileInfo, err error) error {
						if err != nil {
							log.Printf("can't stat file %s: %v\n", path, err)
							return nil
						}
						if !fi.IsDir() {
							return nil
						}
						if strings.HasPrefix(fi.Name(), ".") || strings.HasPrefix(fi.Name(), "_") || fi.Name() == "testdata" {
							return filepath.SkipDir
						}
						//if fi.Name() == "vendor" { // THINK.
						//	return filepath.SkipDir
						//}
						// Determine repo root. This is potentially somewhat slow.
						vcsCmd, root, err := vcs.FromDir(path, srcRoot)
						if err != nil {
							// Directory not under VCS.
							return nil
						}
						pipeline.AddRepo(Repo{Path: path, Root: root, VCS: vcsCmd})
						packages++
						return filepath.SkipDir // No need to descend inside repositories.
					})
				}
			}
			pipeline.Done()
			fmt.Printf("%v packages.\n", packages)
		}()
		updater = repo.GopathUpdater{}
	case *stdinFlag:
		fmt.Println("Reading the list of newline separated Go packages from stdin.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			br := bufio.NewReader(os.Stdin)
			packages := 0
			for line, err := br.ReadString('\n'); err == nil; line, err = br.ReadString('\n') {
				importPath := line[:len(line)-1] // Trim last newline.
				pipeline.Add(importPath)
				packages++
			}
			pipeline.Done()
			fmt.Printf("%v packages.\n", packages)
		}()
		updater = repo.GopathUpdater{}
	case *godepsFlag != "":
		fmt.Println("Reading the list of Go packages from Godeps.json file:", *godepsFlag)
		g, err := readGodeps(*godepsFlag)
		if err != nil {
			// TODO: Handle errors more gracefully.
			log.Fatalln("readGodeps:", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range g.Deps {
				pipeline.AddRevision(dependency.ImportPath, dependency.Rev)
			}
			pipeline.Done()
			fmt.Println("loadGoPackagesFromGodeps done")
		}()
		updater = nil
	case *govendorFlag != "":
		fmt.Println("Reading the list of Go packages from vendor.json file:", *govendorFlag)
		v, err := readGovendor(*govendorFlag)
		if err != nil {
			// TODO: Handle errors more gracefully.
			log.Fatalln("readGovendor:", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range v.Package {
				pipeline.AddRevision(dependency.Path, dependency.Revision)
			}
			pipeline.Done()
			fmt.Println("loadGoPackagesFromGovendor done")
		}()
		updater = nil
	}

	err := loadTemplates()
	if err != nil {
		log.Fatalln("loadTemplates:", err)
	}

	http.HandleFunc("/index.html", mainHandler)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/assets/", gzip_file_server.New(assets))
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

	switch production {
	case true:
		// Open a browser tab and navigate to the main page.
		go u4.Open("http://" + *httpFlag + "/index.html")
	case false:
		updater = repo.MockUpdater{}
	}

	fmt.Println("Go Package Store server is running at http://" + *httpFlag + "/index.html.")

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
