// Go Package Store displays updates for the Go packages in your GOPATH.
package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shurcooL/Go-Package-Store/presenter"
	"github.com/shurcooL/go/exp/14"
	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/shurcooL/go/gists/gist7651991"
	"github.com/shurcooL/go/gists/gist7802150"
	"github.com/shurcooL/go/u/u4"
	"github.com/shurcooL/gostatus/status"
	"golang.org/x/net/websocket"
)

func CommonHat(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	io.WriteString(w, `<html>
	<head>
		<title>Go Package Store</title>
		<link href="assets/style.css" rel="stylesheet" type="text/css" />
		<script src="assets/script.js" type="text/javascript"></script>
		<link rel="stylesheet" href="http://cdnjs.cloudflare.com/ajax/libs/octicons/2.1.2/octicons.css">
	</head>
	<body>
		<div style="width: 100%; text-align: center; background-color: hsl(209, 51%, 92%);">
			<span style="background-color: hsl(209, 51%, 88%); padding: 15px; display: inline-block;">Updates</span>
		</div>
		<script type="text/javascript">
			var sock = new WebSocket("ws://`+*httpFlag+`/opened");
			sock.onopen = function () {
				sock.onclose = function() { alert('Go Package Store server disconnected.'); };
			};
		</script>
		<div class="content">`)
}
func CommonTail(w io.Writer) {
	// TODO: Make installed_updates available before all packages finish loading, so that it works when you update a package early.
	io.WriteString(w, `<div id="installed_updates" style="display: none;"><h3 style="text-align: center;">Installed Updates</h3></div>`)
	io.WriteString(w, "</div></body></html>")
}

// ---

// shouldPresentUpdate determines if the given goPackage should be presented as an available update.
// It checks that the Go package is on default branch, does not have a dirty working tree, and does not have the remote revision.
func shouldPresentUpdate(goPackage *gist7480523.GoPackage) bool {
	return status.PlumbingPresenterV2(goPackage)[:3] == "  +" // Ignore stash.
}

// Writes a <div> presentation for an available update.
func WriteRepoHtml(w http.ResponseWriter, repoPresenter presenter.Presenter) {
	err := t.Execute(w, repoPresenter)
	if err != nil {
		log.Println("t.Execute:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// A cached list of Go packages to work with.
var goPackages exp14.GoPackageList

type updateRequest struct {
	importPathPattern string
	resultChan        chan error
}

var updateRequestChan = make(chan updateRequest)

// updateWorker is a sequential updater of Go packages. It does not update them in parallel
// to avoid race conditions or other problems, since `go get -u` does not seem to protect against that.
func updateWorker() {
	for updateRequest := range updateRequestChan {
		cmd := exec.Command("go", "get", "-u", "-d", updateRequest.importPathPattern)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Print(strings.Join(cmd.Args, " "))

		err := cmd.Run()

		// Invalidate cache of the package's local revision, since it's expected to change after `go get -u`.
		gist7802150.MakeUpdated(goPackages)
		for _, goPackage := range goPackages.List() {
			if rootPath := getRootPath(goPackage); rootPath != "" {
				if gist7480523.GetRepoImportPathPattern(rootPath, goPackage.Bpkg.SrcRoot) == updateRequest.importPathPattern {
					//fmt.Println("ExternallyUpdated", updateRequest.importPathPattern)
					gist7802150.ExternallyUpdated(goPackage.Dir.Repo.VcsLocal.GetSources()[1].(gist7802150.DepNode2ManualI))
					break
				}
			}
		}

		updateRequest.resultChan <- err

		fmt.Println("\nDone.")
	}
}

// Handler for update requests.
func updateHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		if *godepsFlag != "" {
			// TODO: Implement updating Godeps packages.
			log.Fatalln("updating Godeps packages isn't supported yet")
		}

		updateRequest := updateRequest{
			importPathPattern: req.PostFormValue("import_path_pattern"),
			resultChan:        make(chan error),
		}
		updateRequestChan <- updateRequest

		err := <-updateRequest.resultChan
		_ = err // Don't do anything about the error for now.
	}
}

// getRootPath returns the root path of the given goPackage.
func getRootPath(goPackage *gist7480523.GoPackage) (rootPath string) {
	if goPackage.Bpkg.Goroot {
		return ""
	}

	goPackage.UpdateVcs()
	if goPackage.Dir.Repo == nil {
		return ""
	} else {
		return goPackage.Dir.Repo.Vcs.RootPath()
	}
}

// Main index page handler.
func mainHandler(w http.ResponseWriter, req *http.Request) {
	if err := loadTemplates(); err != nil {
		fmt.Fprintln(w, "loadTemplates:", err)
		return
	}

	//started := time.Now()

	CommonHat(w)
	defer CommonTail(w)

	io.WriteString(w, `<div id="checking_updates"><h2 style="text-align: center;">Checking for updates...</h2></div>`)
	io.WriteString(w, `<div id="no_updates" style="display: none;"><h2 style="text-align: center;">No Updates Available</h2></div>`)
	defer io.WriteString(w, `<script>document.getElementById("checking_updates").style.display = "none";</script>`)

	flusher := w.(http.Flusher)
	flusher.Flush()

	notifier := w.(http.CloseNotifier)
	go func() {
		<-notifier.CloseNotify()

		//fmt.Println("Exiting, since the HTTP request was cancelled/interrupted.")
		//close(updateRequestChan)
	}()

	//fmt.Printf("Part 1: %v ms.\n", time.Since(started).Seconds()*1000)

	// rootPath -> []*gist7480523.GoPackage
	var goPackagesInRepo = make(map[string][]*gist7480523.GoPackage)

	gist7802150.MakeUpdated(goPackages)
	//fmt.Printf("Part 1b: %v ms.\n", time.Since(started).Seconds()*1000)
	if false {
		for _, goPackage := range goPackages.List() {
			if rootPath := getRootPath(goPackage); rootPath != "" {
				goPackagesInRepo[rootPath] = append(goPackagesInRepo[rootPath], goPackage)
			}
		}
	} else {
		inChan := make(chan interface{})
		go func() { // This needs to happen in the background because sending input will be blocked on reading output.
			for _, goPackage := range goPackages.List() {
				inChan <- goPackage
			}
			close(inChan)
		}()
		reduceFunc := func(in interface{}) interface{} {
			goPackage := in.(*gist7480523.GoPackage)
			if rootPath := getRootPath(goPackage); rootPath != "" {
				return gist7480523.NewGoPackageRepo(rootPath, []*gist7480523.GoPackage{goPackage})
			}
			return nil
		}
		outChan := gist7651991.GoReduce(inChan, 64, reduceFunc)
		for out := range outChan {
			repo := out.(gist7480523.GoPackageRepo)
			goPackagesInRepo[repo.RootPath()] = append(goPackagesInRepo[repo.RootPath()], repo.GoPackages()[0])
		}
	}

	//goon.DumpExpr(len(goPackages.List()))
	//goon.DumpExpr(len(goPackagesInRepo))

	//fmt.Printf("Part 2: %v ms.\n", time.Since(started).Seconds()*1000)

	updatesAvailable := 0

	inChan := make(chan interface{})
	go func() { // This needs to happen in the background because sending input will be blocked on reading output.
		for rootPath, goPackages := range goPackagesInRepo {
			inChan <- gist7480523.NewGoPackageRepo(rootPath, goPackages)
		}
		close(inChan)
	}()
	reduceFunc := func(in interface{}) interface{} {
		repo := in.(gist7480523.GoPackageRepo)

		goPackage := repo.GoPackages()[0]
		goPackage.UpdateVcsFields()

		if !shouldPresentUpdate(goPackage) {
			return nil
		}
		repoPresenter := presenter.New(&repo)
		return repoPresenter
	}
	outChan := gist7651991.GoReduce(inChan, 8, reduceFunc)

	for out := range outChan {
		//started2 := time.Now()

		repoPresenter := out.(presenter.Presenter)

		updatesAvailable++
		WriteRepoHtml(w, repoPresenter)

		flusher.Flush()

		//fmt.Printf("Part 2b: %v ms.\n", time.Since(started2).Seconds()*1000)
	}

	if updatesAvailable == 0 {
		io.WriteString(w, `<script>document.getElementById("no_updates").style.display = "";</script>`)
	}

	//fmt.Printf("Part 3: %v ms.\n", time.Since(started).Seconds()*1000)
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
	const filename = "./assets/repo.html.tmpl"

	var err error
	t, err = template.ParseFiles(filename)
	return err
}

var httpFlag = flag.String("http", "localhost:7043", "Listen for HTTP connections on this address.")
var stdinFlag = flag.Bool("stdin", false, "Read the list of newline separated Go packages from stdin.")
var godepsFlag = flag.String("godeps", "", "Read the list of Go packages from the specified Godeps file.")

func usage() {
	fmt.Fprint(os.Stderr, "Usage: Go-Package-Store [flags]\n")
	fmt.Fprint(os.Stderr, "       [newline separated packages] | Go-Package-Store --stdin [flags]\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, `
Examples:
  # Check for updates for all Go packages in GOPATH.
  Go-Package-Store

  # Show updates for all dependencies (recursive) of package in cur working dir.
  go list -f '{{join .Deps "\n"}}' . | Go-Package-Store --stdin
`)
	os.Exit(2)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = usage
	flag.Parse()

	switch {
	default:
		fmt.Println("Using all Go packages in GOPATH.")
		goPackages = &exp14.GoPackages{SkipGoroot: true} // All Go packages in GOPATH (not including GOROOT).
	case *stdinFlag:
		fmt.Println("Reading the list of newline separated Go packages from stdin.")
		goPackages = &exp14.GoPackagesFromReader{Reader: os.Stdin}
	case *godepsFlag != "":
		fmt.Println("Reading the list of Go packages from Godeps file:", *godepsFlag)
		goPackages = NewGoPackagesFromGodeps(*godepsFlag)
	}

	// Set the working directory to the root of the package, so that its assets folder can be used.
	{
		goPackageStore := gist7480523.GoPackageFromImportPath("github.com/shurcooL/Go-Package-Store")
		if goPackageStore == nil {
			log.Fatalln("Unable to find github.com/shurcooL/Go-Package-Store package in your GOPATH, it's needed to load assets.")
		}

		err := os.Chdir(goPackageStore.Bpkg.Dir)
		if err != nil {
			log.Panicln("os.Chdir:", err)
		}
	}

	err := loadTemplates()
	if err != nil {
		log.Fatalln("loadTemplates:", err)
	}

	//goon.DumpExpr(os.Getwd())
	//goon.DumpExpr(os.Getenv("PATH"), os.Getenv("GOPATH"))

	http.HandleFunc("/index.html", mainHandler)
	http.HandleFunc("/-/update", updateHandler)
	http.Handle("/favicon.ico/", http.NotFoundHandler())
	http.Handle("/assets/", http.FileServer(http.Dir(".")))
	http.Handle("/opened", websocket.Handler(openedHandler)) // Exit server when client tab is closed.
	go updateWorker()

	// Open a browser tab and navigate to the main page.
	u4.Open("http://" + *httpFlag + "/index.html")

	fmt.Println("Go Package Store server is running at http://" + *httpFlag + "/index.html.")

	err = http.ListenAndServe(*httpFlag, nil)
	if err != nil {
		panic(err)
	}
}
