package main

import (
	"fmt"
	"io"
	"net/http"
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

	importPath := r.URL.Path[1:]

	if goPackage := GoPackageFromImportPath(importPath); goPackage != nil {
		doStuffWithPackage(w, goPackage)
	}
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

	fmt.Println("GithubComparison) Update() {, err:", this.err)
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

func presentGithubHtml(w io.Writer, goPackage *GoPackage) {
	importPath := goPackage.Bpkg.ImportPath
	rootPath := goPackage.Vcs.VcsState.Vcs.RootPath()

	comparison, ok := githubComparisons[rootPath]
	if !ok {
		comparison = NewGithubComparison(importPath, goPackage.Vcs.VcsState.VcsLocal, goPackage.Vcs.VcsState.VcsRemote)
		githubComparisons[rootPath] = comparison
	}

	if MakeUpdated(comparison); comparison.err != nil {
		fmt.Fprintln(w, "couldn't compare:", comparison.err)
	} else {
		GenerateGithubHtml(w, goPackage, comparison.cc)
	}
}

func GenerateGithubHtml(w io.Writer, goPackage *GoPackage, cc *github.CommitsComparison) {
	//goon.DumpExpr(goPackage, cc)

	importPath := goPackage.Bpkg.ImportPath

	fmt.Fprintf(w, `<h3>%s</h3>`, importPath)

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

func doStuffWithPackage(w io.Writer, goPackage *GoPackage) {
	if goPackage.Standard {
		return
	}

	goPackage.UpdateVcs()
	if goPackage.Vcs.VcsState == nil {
		return
	}

	goPackage.UpdateVcsFields()
	if shouldShow(goPackage) == false {
		return
	}
	if shouldPresentGithub(goPackage) {
		presentGithubHtml(w, goPackage)
	} /*else {
		io.WriteString(w, "<p>"+presenter(goPackage)+"</p>")
	}*/
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
	commonHat(w)
	defer commonTail(w)

	flusher := w.(http.Flusher)

	MakeUpdated(goPackages)
	for _, goPackage := range goPackages.Entries {
		doStuffWithPackage(w, goPackage)
		flusher.Flush()
	}
}

func main() {
	http.HandleFunc("/all", mainHandler)
	http.HandleFunc("/-/update", updateHandler)
	http.HandleFunc("/", debugHandler)

	err := http.ListenAndServe(":8080", nil)
	CheckError(err)
}
