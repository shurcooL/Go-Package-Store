package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"strings"

	. "gist.github.com/5286084.git"
	. "gist.github.com/7480523.git"
	. "gist.github.com/7802150.git"

	//. "gist.github.com/7519227.git"
	"github.com/google/go-github/github"
	"github.com/shurcooL/go/exp/13"
	"github.com/shurcooL/go/exp/14"
)

//var presenter GoPackageStringer = status.PorcelainPresenter

var shouldShow = func(goPackage *GoPackage) bool {
	// Check for notable status
	return goPackage.Vcs.VcsState != nil &&
		(goPackage.Vcs.VcsState.VcsLocal.LocalBranch != goPackage.Vcs.VcsState.Vcs.GetDefaultBranch() ||
			goPackage.Vcs.VcsState.VcsLocal.Status != "" ||
			goPackage.Vcs.VcsState.VcsLocal.LocalRev != goPackage.Vcs.VcsState.VcsRemote.RemoteRev)
}

var gh = github.NewClient(nil)

func commonHat(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=us-ascii")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// TODO: Serve and use own css
	io.WriteString(w, `<html><head><title>Go Package Store</title></head><body>`)
}
func commonTail(w http.ResponseWriter) {
	io.WriteString(w, "</body></html>")
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	commonHat(w)
	defer commonTail(w)

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

func shouldPresentGithub(goPackage *GoPackage) bool {
	return strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/") &&
		goPackage.Vcs.VcsState != nil &&
		goPackage.Vcs.VcsState.VcsLocal.LocalBranch == goPackage.Vcs.VcsState.Vcs.GetDefaultBranch() &&
		goPackage.Vcs.VcsState.VcsLocal.Status == "" &&
		goPackage.Vcs.VcsState.VcsLocal.LocalRev != goPackage.Vcs.VcsState.VcsRemote.RemoteRev
}

func updateGithubHtml(goPackage *GoPackage) (rootPath string) {
	importPath := goPackage.Bpkg.ImportPath
	rootPath = goPackage.Vcs.VcsState.Vcs.RootPath()

	comparison, ok := githubComparisons[rootPath]
	if !ok {
		comparison = NewGithubComparison(importPath, goPackage.Vcs.VcsState.VcsLocal, goPackage.Vcs.VcsState.VcsRemote)
		githubComparisons[rootPath] = comparison
	}

	MakeUpdated(comparison)

	return rootPath
}

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
		fmt.Fprintf(w, `<h3>%s and <span title="%s">%d more</span></h3>`, importPath, strings.Join(importPaths[1:], "\n"), len(goPackages)-1)
	}

	if cc.BaseCommit != nil && cc.BaseCommit.Author != nil && cc.BaseCommit.Author.AvatarURL != nil {
		// TODO: Factor out styles into css
		fmt.Fprintf(w, `<img style="float: left; border-radius: 4px;" src="%s" width="36" height="36">`, *cc.BaseCommit.Author.AvatarURL)
	}

	// TODO: Make the forn name unique, because there'll be many on same page
	// TODO: Factor out styles into css
	fmt.Fprint(w, `<div style="float: right;">`)
	fmt.Fprintf(w, `<form style="display: none;" name="x-update" method="POST" action="/-/update"><input type="hidden" name="import_path" value="%s"></form>`, importPath)
	fmt.Fprintf(w, `<a href="javascript:document.getElementsByName('x-update')[0].submit();" title="%s">Update</a>`, fmt.Sprintf("go get -u -d %s", importPath))
	fmt.Fprint(w, `</div>`)

	// TODO: Factor out styles into css
	// HACK: Manually aligned to the left of the image, this should be done via proper html layout
	fmt.Fprint(w, `<div style="padding-left: 36px;">`)
	fmt.Fprint(w, `<ol>`)

	for index := range cc.Commits {
		repositoryCommit := cc.Commits[len(cc.Commits)-1-index]
		if repositoryCommit.Commit != nil && repositoryCommit.Commit.Message != nil {
			fmt.Fprint(w, "<li>")
			fmt.Fprint(w, *repositoryCommit.Commit.Message)
			fmt.Fprint(w, "</li>")
		}
	}

	fmt.Fprint(w, `</ol>`)
	fmt.Fprint(w, `</div>`)
}

func doStuffWithPackage(goPackage *GoPackage) (rootPath string, ok bool) {
	if goPackage.Standard {
		return "", false
	}

	goPackage.UpdateVcs()
	if goPackage.Vcs.VcsState == nil {
		return "", false
	}

	goPackage.UpdateVcsFields()
	if shouldShow(goPackage) == false {
		return "", false
	}
	if shouldPresentGithub(goPackage) {
		rootPath := updateGithubHtml(goPackage)
		return rootPath, true
	} /*else {
		io.WriteString(w, "<p>"+presenter(goPackage)+"</p>")
	}*/
	return "", false
}

var goPackages = &exp14.GoPackages{}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		importPath := r.PostFormValue("import_path")

		// TODO: Activate
		fmt.Println("go", "get", "-u", "-d", importPath)
		_ = exec.Command("go", "get", "-u", "-d", importPath)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	b, err := httputil.DumpRequest(r, false)
	CheckError(err)
	fmt.Println(string(b))

	commonHat(w)
	defer commonTail(w)

	flusher := w.(http.Flusher)

	// rootPath -> []*GoPackage
	var x = make(map[string][]*GoPackage)

	MakeUpdated(goPackages)
	for _, goPackage := range goPackages.Entries {
		if rootPath, ok := doStuffWithPackage(goPackage); ok {
			x[rootPath] = append(x[rootPath], goPackage)
		}
	}

	for rootPath, goPackages := range x {
		comparison := githubComparisons[rootPath]
		if comparison.err != nil {
			fmt.Fprintln(w, "couldn't compare:", comparison.err)
		} else {
			GenerateGithubHtml(w, goPackages, comparison.cc)
		}

		flusher.Flush()
	}
}

func main() {
	http.HandleFunc("/all", mainHandler)
	http.HandleFunc("/-/update", updateHandler)
	//http.HandleFunc("/debug", debugHandler)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	err := http.ListenAndServe(":8080", nil)
	CheckError(err)
}
