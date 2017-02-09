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
	"github.com/shurcooL/Go-Package-Store"
	"github.com/shurcooL/Go-Package-Store/assets"
	gpscomponent "github.com/shurcooL/Go-Package-Store/component"
	"github.com/shurcooL/Go-Package-Store/presenter/github"
	"github.com/shurcooL/Go-Package-Store/presenter/gitiles"
	"github.com/shurcooL/Go-Package-Store/updater"
	"github.com/shurcooL/Go-Package-Store/workspace"
	"github.com/shurcooL/go/open"
	"github.com/shurcooL/go/ospath"
	"github.com/shurcooL/htmlg"
	"github.com/shurcooL/httperror"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"github.com/shurcooL/httpgzip"
	"golang.org/x/net/websocket"
	"golang.org/x/oauth2"
)

var c = struct {
	pipeline *workspace.Pipeline

	updateHandler *updateHandler
}{updateHandler: &updateHandler{updateRequests: make(chan updateRequest)}}

// mainHandler is the handler for the index page.
func mainHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		httperror.HandleMethod(w, httperror.Method{Allowed: []string{"GET"}})
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

	for repoPresentation := range c.pipeline.RepoPresentations() {
		if !repoPresentation.Updated {
			updatesAvailable++
		}

		if repoPresentation.Updated && !wroteInstalledUpdatesHeader {
			// Make 'Installed Updates' header visible now.
			io.WriteString(w, `<div id="installed_updates"><h3 style="text-align: center;">Installed Updates</h3></div>`)

			wroteInstalledUpdatesHeader = true
		}

		err := t.ExecuteTemplate(w, "repo.html.tmpl", repoPresentation)
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
	//close(updateRequests)
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
		"updateSupported": func() bool { return c.updateHandler.updater != nil },

		"render": func(c htmlg.Component) template.HTML { return htmlg.Render(c.Render()...) },
		"change": func(c gps.Change) htmlg.Component {
			return gpscomponent.Change{
				Message:  c.Message,
				URL:      string(c.URL),
				Comments: gpscomponent.Comments{Count: c.Comments.Count, URL: string(c.Comments.URL)},
			}
		},
		"comments": func(c gps.Comments) htmlg.Component { return gpscomponent.Comments{Count: c.Count, URL: string(c.URL)} },
		"commitID": func(commitID string) htmlg.Component { return gpscomponent.CommitID{ID: commitID} },
	})
	t, err = vfstemplate.ParseGlob(assets.Assets, t, "/assets/*.tmpl")
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

  # Show updates for all golang.org/x/... packages.
  go list golang.org/x/... | Go-Package-Store -stdin

  # Show updates for all dependencies listed in vendor.json file.
  Go-Package-Store -govendor=/path/to/repo/vendor/vendor.json

  # Show updates for all Go packages vendored using git-subrepo
  # in the specified vendor directory.
  Go-Package-Store -git-subrepo=/path/to/repo/vendor
`)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	log.SetFlags(0)

	c.pipeline = workspace.NewPipeline(wd)

	// If we can have access to a cache directory on this system, use it for
	// caching HTTP requests of presenters.
	cacheDir, err := ospath.CacheDir("github.com/shurcooL/Go-Package-Store")
	if err != nil {
		log.Println("skipping persistent on-disk caching, because unable to acquire a cache dir:", err)
		cacheDir = ""
	}

	// Register GitHub presenter.
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

		c.pipeline.RegisterPresenter(github.NewPresenter(&http.Client{Transport: transport}))
	}

	// Register Gitiles presenter.
	{
		var transport http.RoundTripper

		if cacheDir != "" {
			transport = &httpcache.Transport{
				Transport:           transport,
				Cache:               diskcache.New(filepath.Join(cacheDir, "gitiles-presenter")),
				MarkCachedResponses: true,
			}
		}

		c.pipeline.RegisterPresenter(gitiles.NewPresenter(&http.Client{Transport: transport}))
	}

	switch {
	case !production:
		fmt.Println("Using no real packages (hit /mock.html endpoint for mocks).")
		c.pipeline.Done()
		c.updateHandler.updater = updater.Mock{}
	default:
		fmt.Println("Using all Go packages in GOPATH.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			forEachRepository(func(r workspace.LocalRepo) {
				c.pipeline.AddRepository(r)
			})
			c.pipeline.Done()
		}()
		c.updateHandler.updater = updater.Gopath{}
	case *stdinFlag:
		fmt.Println("Reading the list of newline separated Go packages from stdin.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			br := bufio.NewReader(os.Stdin)
			for line, err := br.ReadString('\n'); err == nil; line, err = br.ReadString('\n') {
				importPath := line[:len(line)-1] // Trim last newline.
				c.pipeline.AddImportPath(importPath)
			}
			c.pipeline.Done()
		}()
		c.updateHandler.updater = updater.Gopath{}
	case *godepsFlag != "":
		fmt.Println("Reading the list of Go packages from Godeps.json file:", *godepsFlag)
		g, err := readGodeps(*godepsFlag)
		if err != nil {
			log.Fatalln("Failed to read Godeps.json file", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range g.Deps {
				c.pipeline.AddRevision(dependency.ImportPath, dependency.Rev)
			}
			c.pipeline.Done()
		}()
		c.updateHandler.updater = nil
	case *govendorFlag != "":
		fmt.Println("Reading the list of Go packages from vendor.json file:", *govendorFlag)
		v, err := readGovendor(*govendorFlag)
		if err != nil {
			log.Fatalln("Failed to read vendor.json file:", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range v.Package {
				c.pipeline.AddRevision(dependency.Path, dependency.Revision)
			}
			c.pipeline.Done()
		}()
		// TODO: Consider setting a better directory for govendor command than current working directory.
		//       Perhaps the parent directory of vendor.json file?
		if gu, err := updater.NewGovendor(""); err == nil {
			c.updateHandler.updater = gu
		} else {
			log.Println("govendor updater is not available:", err)
		}
	case *gitSubrepoFlag != "":
		if _, err := exec.LookPath("git"); err != nil {
			log.Fatalln(fmt.Errorf("git binary is required, but not available: %v", err))
		}
		fmt.Println("Using Go packages vendored using git-subrepo in the specified vendor directory.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			err := forEachGitSubrepo(*gitSubrepoFlag, func(s workspace.Subrepo) {
				c.pipeline.AddSubrepo(s)
			})
			if err != nil {
				log.Println("warning: there was problem iterating over subrepos:", err)
			}
			c.pipeline.Done()
		}()
		c.updateHandler.updater = nil // An updater for this can easily be added by anyone who uses this style of vendoring.
	}

	err = loadTemplates()
	if err != nil {
		log.Fatalln("loadTemplates:", err)
	}

	http.HandleFunc("/index.html", mainHandler)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	fileServer := httpgzip.FileServer(assets.Assets, httpgzip.FileServerOptions{ServeError: httpgzip.Detailed})
	http.Handle("/assets/", fileServer)
	http.Handle("/assets/octicons/", http.StripPrefix("/assets", fileServer))
	http.Handle("/opened", websocket.Handler(openedHandler)) // Exit server when client tab is closed.
	if c.updateHandler.updater != nil {
		http.Handle("/-/update", c.updateHandler)
		go c.updateHandler.Worker()
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
