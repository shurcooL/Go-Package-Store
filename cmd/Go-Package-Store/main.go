// Go Package Store displays updates for the Go packages in your GOPATH.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	gps "github.com/shurcooL/Go-Package-Store"
	"github.com/shurcooL/Go-Package-Store/assets"
	"github.com/shurcooL/Go-Package-Store/presenter/github"
	"github.com/shurcooL/Go-Package-Store/presenter/gitiles"
	"github.com/shurcooL/Go-Package-Store/updater"
	"github.com/shurcooL/Go-Package-Store/workspace"
	"github.com/shurcooL/go/browser"
	"github.com/shurcooL/go/ospath"
	"github.com/shurcooL/httpgzip"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
)

var (
	httpFlag       = flag.String("http", "localhost:7043", "Listen for HTTP connections on this address.")
	stdinFlag      = flag.Bool("stdin", false, "Read the list of newline separated Go packages from stdin.")
	depFlag        = flag.String("dep", "", "Determine the list of Go packages from the specified Gopkg.toml file.")
	godepsFlag     = flag.String("godeps", "", "Read the list of Go packages from the specified Godeps.json file.")
	govendorFlag   = flag.String("govendor", "", "Read the list of Go packages from the specified vendor.json file.")
	gitSubrepoFlag = flag.String("git-subrepo", "", "Look for Go packages vendored using git-subrepo in the specified vendor directory.")
)

func usage() {
	fmt.Fprint(os.Stderr, "Usage: Go-Package-Store [flags]\n")
	fmt.Fprint(os.Stderr, "       [newline separated packages] | Go-Package-Store -stdin [flags]\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, `
GitHub Access Token:
  To display updates for private repositories on GitHub, or when
  you've exceeded the unauthenticated rate limit, you can provide
  a GitHub access token for Go Package Store to use via the
  GO_PACKAGE_STORE_GITHUB_TOKEN environment variable.

Examples:
  # Check for updates for all Go packages in GOPATH.
  Go-Package-Store

  # Show updates for all golang.org/x/... packages.
  go list golang.org/x/... | Go-Package-Store -stdin

  # Show updates for all dependencies within Gopkg.toml constraints.
  Go-Package-Store -dep=/path/to/repo/Gopkg.toml

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
	registerPresenters(c.pipeline)
	c.updater = populatePipelineAndCreateUpdater(c.pipeline)
	if c.updater != nil {
		updateWorker := newUpdateWorker(c.updater)
		updateWorker.Start()
		http.Handle("/api/update", errorHandler(updateWorker.Handler))
	}
	http.Handle("/api/updates", errorHandler(updatesHandler))
	http.Handle("/updates", errorHandler(indexHandler))
	assetsFS := httpgzip.FileServer(assets.Assets, httpgzip.FileServerOptions{ServeError: httpgzip.Detailed})
	http.Handle("/assets/", assetsFS)
	http.Handle("/frontend.js", assetsFS)
	fontsFS := httpgzip.FileServer(assets.Fonts, httpgzip.FileServerOptions{ServeError: httpgzip.Detailed})
	http.Handle("/assets/fonts/", http.StripPrefix("/assets/fonts", fontsFS))

	// Start listening first.
	listener, err := net.Listen("tcp", *httpFlag)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to listen on %q: %v", *httpFlag, err))
	}

	if production {
		// Open a browser tab and navigate to the main page.
		go browser.Open("http://" + *httpFlag + "/updates")
	}

	fmt.Println("Go Package Store server is running at http://" + *httpFlag + "/updates.")

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

// c is a global context.
var c = struct {
	pipeline *workspace.Pipeline

	// updater is set based on the source of Go packages. If nil, it means
	// we don't have support to update Go packages from the current source.
	// It's used to update repos in the backend, and if set to nil, to disable
	// the frontend UI for updating packages.
	updater gps.Updater
}{}

func registerPresenters(pipeline *workspace.Pipeline) {
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

		pipeline.RegisterPresenter(github.NewPresenter(&http.Client{Transport: transport}))
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

		pipeline.RegisterPresenter(gitiles.NewPresenter(&http.Client{Transport: transport}))
	}
}

func populatePipelineAndCreateUpdater(pipeline *workspace.Pipeline) gps.Updater {
	switch {
	case !production:
		fmt.Println("Using no real packages (hit /mock.html or /component.html endpoint for mocks).")
		pipeline.Done()
		return updater.Mock{}
	default:
		// Get the GOMOD value, use it to determine if godoc is being invoked in module mode.
		goModFile, err := goMod()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to determine go env GOMOD value: %v", err)
			goModFile = "" // Fall back to GOPATH mode.
		}

		if goModFile != "" {
			fmt.Printf("using module mode; GOMOD=%s\n", goModFile)

			mods, err := buildList(goModFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to determine the build list of the main module: %v", err)
				os.Exit(1)
			}

			// Bind module trees into Go root.
			for _, m := range mods {
				if m.Update == nil || m.Indirect {
					continue
				}
				pipeline.AddModule(workspace.Module{
					Path:     m.Path,
					Version:  m.Version,
					Update:   m.Update.Version,
					Indirect: m.Indirect,
				})
			}
			pipeline.Done()

			return updater.Mock{}
		} else {
			fmt.Println("using GOPATH mode")

			fmt.Println("Using all Go packages in GOPATH.")
			go func() { // This needs to happen in the background because sending input will be blocked on processing.
				forEachRepository(func(r workspace.LocalRepo) {
					pipeline.AddRepository(r)
				})
				pipeline.Done()
			}()
			return updater.Gopath{}
		}
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
		return updater.Gopath{}
	case *depFlag != "":
		// Check dep binary exists in PATH.
		if _, err := exec.LookPath("dep"); err != nil {
			log.Fatalln(fmt.Errorf("dep binary is required, but not available: %v", err))
		}
		fmt.Println("Determining the list of Go packages from Gopkg.toml file:", *depFlag)
		dir, err := depDir(*depFlag)
		if err != nil {
			log.Fatalln(err)
		}
		go func() {
			// Running dep status is pretty slow, because it tries to figure out remote updates
			// that are within Gopkg.toml constraints. So run it in background.
			dependencies, err := runDepStatus(dir)
			if err != nil {
				log.Println("failed to run dep status on Gopkg.toml file:", err)
				dependencies = nil
			}

			// This needs to happen in the background because sending input will be blocked on processing.
			for _, d := range dependencies {
				pipeline.AddRevisionLatest(d.ProjectRoot, d.Revision, d.Latest)
			}
			pipeline.Done()
		}()
		return updater.Dep{Dir: dir}
	case *godepsFlag != "":
		fmt.Println("Reading the list of Go packages from Godeps.json file:", *godepsFlag)
		g, err := readGodeps(*godepsFlag)
		if err != nil {
			log.Fatalln("failed to read Godeps.json file", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range g.Deps {
				pipeline.AddRevision(dependency.ImportPath, dependency.Rev)
			}
			pipeline.Done()
		}()
		return nil
	case *govendorFlag != "":
		fmt.Println("Reading the list of Go packages from vendor.json file:", *govendorFlag)
		v, err := readGovendor(*govendorFlag)
		if err != nil {
			log.Fatalln("failed to read vendor.json file:", err)
		}
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			for _, dependency := range v.Package {
				pipeline.AddRevision(dependency.Path, dependency.Revision)
			}
			pipeline.Done()
		}()
		// TODO: Consider setting a better directory for govendor command than current working directory.
		//       Perhaps the parent directory of vendor.json file?
		gu, err := updater.NewGovendor("")
		if err != nil {
			log.Println("govendor updater is not available:", err)
			gu = nil
		}
		return gu
	case *gitSubrepoFlag != "":
		if _, err := exec.LookPath("git"); err != nil {
			log.Fatalln(fmt.Errorf("git binary is required, but not available: %v", err))
		}
		fmt.Println("Using Go packages vendored using git-subrepo in the specified vendor directory.")
		go func() { // This needs to happen in the background because sending input will be blocked on processing.
			err := forEachGitSubrepo(*gitSubrepoFlag, func(s workspace.Subrepo) {
				pipeline.AddSubrepo(s)
			})
			if err != nil {
				log.Println("warning: there was problem iterating over subrepos:", err)
			}
			pipeline.Done()
		}()
		return nil // An updater for this can easily be added by anyone who uses this style of vendoring.
	}
}

// goMod returns the go env GOMOD value in the current directory
// by invoking the go command.
//
// GOMOD is documented at https://golang.org/cmd/go/#hdr-Environment_variables:
//
// 	The absolute path to the go.mod of the main module,
// 	or the empty string if not using modules.
//
func goMod() (string, error) {
	out, err := exec.Command("go", "env", "-json", "GOMOD").Output()
	if ee := (*exec.ExitError)(nil); xerrors.As(err, &ee) {
		return "", fmt.Errorf("go command exited unsuccessfully: %v\n%s", ee.ProcessState.String(), ee.Stderr)
	} else if err != nil {
		return "", err
	}
	var env struct {
		GoMod string
	}
	err = json.Unmarshal(out, &env)
	if err != nil {
		return "", err
	}
	return env.GoMod, nil
}

// buildList determines the build list in the current directory
// by invoking the go command. It should only be used when operating
// in module mode.
//
// See https://golang.org/cmd/go/#hdr-The_main_module_and_the_build_list.
func buildList(goMod string) ([]mod, error) {
	if goMod == os.DevNull {
		// Empty build list.
		return nil, nil
	}

	out, err := exec.Command("go", "list", "-m", "-u", "-json", "all").Output()
	if ee := (*exec.ExitError)(nil); xerrors.As(err, &ee) {
		return nil, fmt.Errorf("go command exited unsuccessfully: %v\n%s", ee.ProcessState.String(), ee.Stderr)
	} else if err != nil {
		return nil, err
	}
	var mods []mod
	for dec := json.NewDecoder(bytes.NewReader(out)); ; {
		var m mod
		err := dec.Decode(&m)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	return mods, nil
}

type mod struct {
	Path     string     // Module path.
	Version  string     // Module version.
	Time     *time.Time // Time version was created.
	Update   *mod       // Available update, if any.
	Indirect bool       // Is this module only an indirect dependency of main module?
}

// wd is current working directory at process start.
var wd = func() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln("os.Getwd:", err)
	}
	return wd
}()
