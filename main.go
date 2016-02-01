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

	"github.com/shurcooL/Go-Package-Store/pkg"
	_ "github.com/shurcooL/Go-Package-Store/presenter/github"
	"github.com/shurcooL/Go-Package-Store/repo"
	"github.com/shurcooL/go/gzip_file_server"
	"github.com/shurcooL/go/u/u4"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"golang.org/x/net/websocket"
)

// shouldPresentUpdate determines if the given goPackage should be presented as an available update.
// It checks that the Go package is on default branch, does not have a dirty working tree, and does not have the remote revision.
func shouldPresentUpdate(repo *pkg.Repo) bool {
	// TODO: Replicate the previous behavior fully, then remove this commented out code:
	//return status.PlumbingPresenterV2(goPackage)[:3] == "  +" // Ignore stash.

	if repo.RemoteURL == "" || repo.Local.Revision == "" || repo.Remote.Revision == "" {
		return false
	}

	if repo.VCS != nil {
		if b, err := repo.VCS.Branch(repo.Path); err != nil || b != repo.VCS.DefaultBranch() {
			return false
		}
		if s, err := repo.VCS.Status(repo.Path); err != nil || s != "" {
			return false
		}
		if c, err := repo.VCS.Contains(repo.Path, repo.Remote.Revision); err != nil || c {
			return false
		}
	}

	return repo.Local.Revision != repo.Remote.Revision
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
	_ = err // TODO: Maybe display error in frontend. For now, don't do anything.
}

// mainHandler is the handler for the index page.
func mainHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method should be GET", http.StatusMethodNotAllowed)
		return
	}

	if err := loadTemplates(); err != nil {
		fmt.Fprintln(w, "loadTemplates:", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
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

	updatesAvailable := 0

	for presented := range pipeline.Presented() {
		updatesAvailable++

		err := t.ExecuteTemplate(w, "repo.html.tmpl", presented)
		if err != nil {
			log.Println("ExecuteTemplate repo.html.tmpl:", err)
			return
		}

		flusher.Flush()
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
		"updateSupported": func() bool { return updater != nil },
		"commitID":        func(commitID string) string { return commitID[:8] },
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
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			forEachRepository(func(r Repo) {
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
				pipeline.Add(importPath)
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
		updater = nil
	}
	if !production {
		updater = repo.MockUpdater{}
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

	if production {
		// Open a browser tab and navigate to the main page.
		go u4.Open("http://" + *httpFlag + "/index.html")
	}

	fmt.Println("Go Package Store server is running at http://" + *httpFlag + "/index.html.")

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
