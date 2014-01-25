package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path"
	"strings"
	"sync"

	. "gist.github.com/5286084.git"
	. "gist.github.com/7480523.git"
	//. "gist.github.com/7519227.git"

	"gist.github.com/8018045.git"
	"github.com/google/go-github/github"
	"github.com/shurcooL/gostatus/status"
)

var _ = github.Bool

// ---

type FlushWriter struct {
	w io.Writer
	f http.Flusher
}

func (fw *FlushWriter) Write(p []byte) (n int, err error) {
	defer fw.f.Flush()
	return fw.w.Write(p)
}

// ---

var presenter GoPackageStringer = status.PorcelainPresenter

var shouldShow = func(goPackage *GoPackage) bool {
	// Check for notable status
	return goPackage.Vcs != nil &&
		(goPackage.LocalBranch != goPackage.Vcs.GetDefaultBranch() ||
			goPackage.Status != "" ||
			goPackage.Local != goPackage.Remote)
}

var lock sync.Mutex
var checkedRepos = make(map[string]bool)

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

func shouldPresentGithub(goPackage *GoPackage) bool {
	return strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/") &&
		goPackage.LocalBranch == goPackage.Vcs.GetDefaultBranch() &&
		goPackage.Status == "" &&
		goPackage.Local != goPackage.Remote
}

func presentGithubHtml(w io.Writer, goPackage *GoPackage) {
	importPath := goPackage.Bpkg.ImportPath
	importPathElements := strings.Split(importPath, "/")
	cc, _, err := gh.Repositories.CompareCommits(importPathElements[1], path.Join(importPathElements[2:]...), goPackage.Local, goPackage.Remote)
	if err != nil {
		fmt.Fprintln(w, "couldn't compare")
		return
	}

	GenerateGithubHtml(w, goPackage, cc)
}

func GenerateGithubHtml(w io.Writer, goPackage *GoPackage, cc *github.CommitsComparison) {
	//goon.DumpExpr(goPackage, cc)

	importPath := goPackage.Bpkg.ImportPath

	fmt.Fprintf(w, `<h3>%s</h3>`, importPath)

	// TODO: Make the forn name unique, because there'll be many on same page
	fmt.Fprintf(w, `<form name="x-update" method="POST" action="/-/update"><input type="hidden" name="import_path" value="%s"></form>`, importPath)
	fmt.Fprintf(w, `<a href="javascript:document.getElementsByName('x-update')[0].submit();" title="%s">Update</a>`, fmt.Sprintf("go get -u -d %s", importPath))

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
}

func doStuffWithPackage(w io.Writer, goPackage *GoPackage) {
	if !goPackage.Standard {
		// HACK: Check that the same repo hasn't already been done
		if goPackage.UpdateVcs(); goPackage.Vcs != nil {
			rootPath := goPackage.Vcs.RootPath()
			lock.Lock()
			if !checkedRepos[rootPath] {
				checkedRepos[rootPath] = true
				lock.Unlock()
			} else {
				lock.Unlock()
				// TODO: Instead of skipping repos that were done, cache their state and reuse it
				return
			}
		}

		goPackage.UpdateVcsFields()
		if shouldShow(goPackage) == false {
			return
		}
		if shouldPresentGithub(goPackage) {
			presentGithubHtml(w, goPackage)
		} else {
			io.WriteString(w, "<p>"+presenter(goPackage)+"</p>")
		}
		return
	}
}

func doStuff(w io.Writer) {
	goPackages := make(chan *GoPackage, 64)

	go gist8018045.GetGoPackages2(goPackages)

	for {
		if goPackage, ok := <-goPackages; ok {
			doStuffWithPackage(w, goPackage)
		} else {
			break
		}
	}
}

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

	fw := &FlushWriter{w: w, f: w.(http.Flusher)}
	doStuff(fw)
}

func main() {
	http.HandleFunc("/all", mainHandler)
	http.HandleFunc("/-/update", updateHandler)
	http.HandleFunc("/", debugHandler)

	err := http.ListenAndServe(":8080", nil)
	CheckError(err)
}
