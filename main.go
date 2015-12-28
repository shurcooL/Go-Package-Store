// Go Package Store displays updates for the Go packages in your GOPATH.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/shurcooL/Go-Package-Store/pkg"
	"github.com/shurcooL/Go-Package-Store/presenter"
	"github.com/shurcooL/Go-Package-Store/repo"
	"github.com/shurcooL/go/gzip_file_server"
	"github.com/shurcooL/go/u/u4"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"golang.org/x/net/websocket"
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

	if repo.Remote.IsContained {
		return false
	}

	if repo.VCS != nil {
		if repo.VCS.GetLocalBranch() != repo.VCS.GetDefaultBranch() {
			return false
		}
		if repo.VCS.GetStatus() != "" {
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
	workspace *goWorkspace = NewGoWorkspace()

	// updater is set based on the source of Go packages. If nil, it means
	// we don't have support to update Go packages from the current source.
	// It's used to update repos in the backend, and to disable the frontend UI
	// for updating packages.
	updater repo.Updater
)

type updateRequest struct {
	importPathPattern string
	resultChan        chan error
}

var updateRequestChan = make(chan updateRequest)

// updateWorker is a sequential updater of Go packages. It does not update them in parallel
// to avoid race conditions or other problems, since `go get -u` does not seem to protect against that.
func updateWorker() {
	for updateRequest := range updateRequestChan {
		err := updater.Update(updateRequest.importPathPattern)
		updateRequest.resultChan <- err
		fmt.Println("\nDone.")
	}
}

// Handler for update requests.
func updateHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}

	updateRequest := updateRequest{
		importPathPattern: req.PostFormValue("import_path_pattern"),
		resultChan:        make(chan error),
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

	for out := range workspace.Out() {
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

func main() {
	flag.Usage = usage
	flag.Parse()

	switch {
	default:
		fmt.Println("Using all Go packages in GOPATH.")
		//goPackages = &exp14.GoPackages{SkipGoroot: true} // All Go packages in GOPATH (not including GOROOT).
		//updater = repo.GopathUpdater{GoPackages: goPackages}
	case *stdinFlag:
		fmt.Println("Reading the list of newline separated Go packages from stdin.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			br := bufio.NewReader(os.Stdin)
			packages := 0
			for line, err := br.ReadString('\n'); err == nil; line, err = br.ReadString('\n') {
				importPath := line[:len(line)-1] // Trim last newline.
				workspace.Add(importPath)
				packages++
			}
			workspace.Done()
			fmt.Printf("%v packages.\n", packages)
		}()
		updater = repo.GopathUpdater{GoPackages: workspace.GoPackageList}
	case *godepsFlag != "":
		fmt.Println("Reading the list of Go packages from Godeps.json file:", *godepsFlag)
		g, err := readGodeps(*godepsFlag)
		if err != nil {
			// TODO: Handle errors more gracefully.
			log.Fatalln("readGodeps:", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range g.Deps {
				workspace.AddRevision(dependency.ImportPath, dependency.Rev)
			}
			workspace.Done()
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
				workspace.AddRevision(dependency.Path, dependency.Revision)
			}
			workspace.Done()
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
