package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	. "gist.github.com/5286084.git"
	. "gist.github.com/7480523.git"
	. "gist.github.com/7651991.git"
	. "gist.github.com/7802150.git"

	//. "gist.github.com/7519227.git"
	"github.com/google/go-github/github"
	"github.com/shurcooL/go-goon"
	"github.com/shurcooL/go/exp/13"
	"github.com/shurcooL/go/exp/14"
)

var gh = github.NewClient(nil)

func CommonHat(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=us-ascii")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	io.WriteString(w, `<html><head><title>Go Package Store</title>
<link href="assets/style.css" rel="stylesheet" type="text/css" />
<script src="assets/script.js" type="text/javascript"></script>
</head><body>`)
}
func CommonTail(w http.ResponseWriter) {
	io.WriteString(w, "</body></html>")
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	CommonHat(w)
	defer CommonTail(w)

	/*importPath := r.URL.Path[1:]

	if goPackage := GoPackageFromImportPath(importPath); goPackage != nil {
		doStuffWithPackage(w, goPackage)
	}*/

	/*MakeUpdated(goPackages)
	for _, goPackage := range goPackages.Entries {
		fmt.Fprint(w, goPackage.Bpkg.ImportPath, "<br>")
	}*/

	/*// rootPath -> []*GoPackage
	var x = make(map[string][]*GoPackage)

	MakeUpdated(goPackages)
	for _, goPackage := range goPackages.Entries {
		if rootPath, ok := doStuffWithPackage(goPackage); ok {
			x[rootPath] = append(x[rootPath], goPackage)
		}
	}

	for rootPath, goPackages := range x {
		fmt.Fprint(w, "<b>", rootPath, "</b><br>")
		for _, goPackage := range goPackages {
			fmt.Fprint(w, goPackage.Bpkg.ImportPath, "<br>")
		}
		fmt.Fprint(w, "<br>")
	}*/
}

// ---

type GithubComparison struct {
	importPath string

	cc  *github.CommitsComparison
	err error

	DepNode2
}

func (this *GithubComparison) Update() {
	localRev := this.GetSources()[0].(*exp13.VcsLocal).LocalRev
	remoteRev := this.GetSources()[1].(*exp13.VcsRemote).RemoteRev

	importPathElements := strings.Split(this.importPath, "/")
	this.cc, _, this.err = gh.Repositories.CompareCommits(importPathElements[1], importPathElements[2], localRev, remoteRev)

	//goon.DumpExpr("GithubComparison.Update()", this.importPath, localRev, remoteRev)
	//fmt.Println(this.err)
}

func NewGithubComparison(importPath string, local *exp13.VcsLocal, remote *exp13.VcsRemote) *GithubComparison {
	this := &GithubComparison{importPath: importPath}
	this.AddSources(local, remote)
	return this
}

// rootPath -> *VcsState
var githubComparisons = make(map[string]*GithubComparison)

// ---

func shouldPresentUpdate(goPackage *GoPackage) bool {
	return goPackage.Dir.Repo != nil &&
		goPackage.Dir.Repo.VcsLocal.LocalBranch == goPackage.Dir.Repo.Vcs.GetDefaultBranch() &&
		goPackage.Dir.Repo.VcsLocal.Status == "" &&
		goPackage.Dir.Repo.VcsLocal.LocalRev != goPackage.Dir.Repo.VcsRemote.RemoteRev
}

// TODO: Should really use html/template...
func GenerateGithubHtml(w io.Writer, goPackages []*GoPackage, cc *github.CommitsComparison) {
	//goon.DumpExpr(goPackage, cc)

	var importPaths []string
	for _, goPackage := range goPackages {
		importPaths = append(importPaths, goPackage.Bpkg.ImportPath)
	}

	importPath := goPackages[0].Bpkg.ImportPath

	if len(goPackages) == 1 {
		fmt.Fprintf(w, `<h3>%s</h3>`, importPath)
	} else if len(goPackages) > 1 {
		fmt.Fprintf(w, `<h3>%s <span class="smaller" title="%s">and %d more</span></h3>`, importPath, strings.Join(importPaths[1:], "\n"), len(goPackages)-1)
	}

	if cc.BaseCommit != nil && cc.BaseCommit.Author != nil && cc.BaseCommit.Author.AvatarURL != nil {
		// TODO: Factor out styles into css
		fmt.Fprintf(w, `<img style="float: left; border-radius: 4px;" src="%s" width="36" height="36">`, *cc.BaseCommit.Author.AvatarURL)
	}

	// TODO: Factor out styles into css
	fmt.Fprint(w, `<div style="float: right;">`)
	fmt.Fprintf(w, `<a href="javascript:void(0)" onclick="update_go_package(this);" id="%s" title="%s">Update</a>`, importPath, fmt.Sprintf("go get -u -d %s", importPath))
	fmt.Fprint(w, `</div>`)

	// TODO: Factor out styles into css
	// HACK: Manually aligned to the left of the image, this should be done via proper html layout
	fmt.Fprint(w, `<div style="padding-left: 36px;">`)
	fmt.Fprint(w, `<ol>`)

	for index := range cc.Commits {
		repositoryCommit := cc.Commits[len(cc.Commits)-1-index]
		if repositoryCommit.Commit != nil && repositoryCommit.Commit.Message != nil {
			fmt.Fprint(w, "<li>")
			fmt.Fprint(w, html.EscapeString(*repositoryCommit.Commit.Message))
			fmt.Fprint(w, "</li>")
		}
	}

	fmt.Fprint(w, `</ol>`)
	fmt.Fprint(w, `</div>`)
}

func GenerateGenericHtml(w io.Writer, goPackages []*GoPackage) {
	var importPaths []string
	for _, goPackage := range goPackages {
		importPaths = append(importPaths, goPackage.Bpkg.ImportPath)
	}

	importPath := goPackages[0].Bpkg.ImportPath

	if len(goPackages) == 1 {
		fmt.Fprintf(w, `<h3>%s</h3>`, importPath)
	} else if len(goPackages) > 1 {
		fmt.Fprintf(w, `<h3>%s <span class="smaller" title="%s">and %d more</span></h3>`, importPath, strings.Join(importPaths[1:], "\n"), len(goPackages)-1)
	}

	// TODO: Factor out styles into css
	fmt.Fprint(w, `<div style="float: right;">`)
	fmt.Fprintf(w, `<a href="javascript:void(0)" onclick="update_go_package(this);" id="%s" title="%s">Update</a>`, importPath, fmt.Sprintf("go get -u -d %s", importPath))
	fmt.Fprint(w, `</div>`)

	fmt.Fprintf(w, `<div>unknown changes</div>`)
}

var goPackages = &exp14.GoPackages{}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		importPath := r.PostFormValue("import_path")

		fmt.Println("go", "get", "-u", "-d", importPath)

		cmd := exec.Command("go", "get", "-u", "-d", importPath)

		out, err := cmd.CombinedOutput()
		goon.DumpExpr(out, err)

		MakeUpdated(goPackages)
		for _, goPackage := range goPackages.Entries {
			if goPackage.Bpkg.ImportPath == importPath {
				ExternallyUpdated(goPackage.Dir.Repo.VcsLocal.GetSources()[1].(DepNode2ManualI))
				break
			}
		}

		fmt.Println("done", importPath)
	}
}

type Repo struct {
	rootPath   string
	goPackages []*GoPackage
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	started := time.Now()

	b, err := httputil.DumpRequest(r, false)
	CheckError(err)
	fmt.Println(string(b))

	CommonHat(w)
	defer CommonTail(w)

	io.WriteString(w, `<div id="checking_updates"><h2 style="text-align: center;">Checking for updates...</h2></div>`)
	defer io.WriteString(w, `<script>document.getElementById("checking_updates").style.display = "none";</script>`)

	flusher := w.(http.Flusher)
	flusher.Flush()

	fmt.Printf("Part 1: %v ms.\n", time.Since(started).Seconds()*1000)

	// rootPath -> []*GoPackage
	var goPackagesInRepo = make(map[string][]*GoPackage)

	// TODO: Use http.CloseNotifier, e.g. https://sourcegraph.com/github.com/donovanhide/eventsource/tree/master/server.go#L70

	getRootPath := func(goPackage *GoPackage) (rootPath string) {
		if goPackage.Standard {
			return ""
		}

		goPackage.UpdateVcs()
		if goPackage.Dir.Repo == nil {
			return ""
		} else {
			return goPackage.Dir.Repo.Vcs.RootPath()
		}
	}

	MakeUpdated(goPackages)
	fmt.Printf("Part 1b: %v ms.\n", time.Since(started).Seconds()*1000)
	for _, goPackage := range goPackages.Entries {
		if rootPath := getRootPath(goPackage); rootPath != "" {
			goPackagesInRepo[rootPath] = append(goPackagesInRepo[rootPath], goPackage)
		}
	}

	fmt.Printf("Part 2: %v ms.\n", time.Since(started).Seconds()*1000)

	updatesAvailable := 0

	reduceFunc := func(in interface{}) interface{} {
		repo := in.(Repo)

		goPackage := repo.goPackages[0]
		goPackage.UpdateVcsFields()

		if !shouldPresentUpdate(goPackage) {
			return nil
		}
		return repo
	}

	inChan := make(chan interface{})
	outChan := GoReduce(inChan, 8, reduceFunc)
	for rootPath, goPackages := range goPackagesInRepo {
		inChan <- Repo{rootPath, goPackages}
	}
	close(inChan)

	for out := range outChan {
		started2 := time.Now()

		repo := out.(Repo)

		goPackage := repo.goPackages[0]

		if strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/") {
			// updateGithubHtml
			comparison, ok := githubComparisons[repo.rootPath]
			if !ok {
				comparison = NewGithubComparison(goPackage.Bpkg.ImportPath, goPackage.Dir.Repo.VcsLocal, goPackage.Dir.Repo.VcsRemote)
				githubComparisons[repo.rootPath] = comparison
			}
			MakeUpdated(comparison)

			if comparison.err != nil {
				fmt.Fprintln(w, "couldn't compare:", comparison.err)
			} else {
				updatesAvailable++
				GenerateGithubHtml(w, repo.goPackages, comparison.cc)
			}
		} else {
			updatesAvailable++
			GenerateGenericHtml(w, repo.goPackages)
		}

		flusher.Flush()

		fmt.Printf("Part 2b: %v ms.\n", time.Since(started2).Seconds()*1000)
	}

	if updatesAvailable == 0 {
		io.WriteString(w, `<div><h2 style="text-align: center;">No Updates Available</h2></div>`)
	}

	fmt.Printf("Part 3: %v ms.\n", time.Since(started).Seconds()*1000)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	goon.DumpExpr(os.Getwd())
	goon.DumpExpr(os.Getenv("PATH"), os.Getenv("GOPATH"))

	http.HandleFunc("/index", mainHandler)
	http.HandleFunc("/-/update", updateHandler)
	//http.HandleFunc("/debug", debugHandler)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/assets/", http.FileServer(http.Dir(".")))

	go func() {
		cmd := exec.Command("open", "http://localhost:7043/index")
		_ = cmd.Run()
	}()

	err := http.ListenAndServe("localhost:7043", nil)
	CheckError(err)
}
