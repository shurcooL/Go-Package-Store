package main

import (
	"fmt"
	"io"
	"net/http"
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

func debugHandler(w http.ResponseWriter, r *http.Request) {
	importPath := r.URL.Path[1:]

	fmt.Fprintln(w, importPath)
	fmt.Fprintln(w)

	if goPackage := GoPackageFromImportPath(importPath); goPackage != nil {
		if strings.HasPrefix(importPath, "github.com/") {
			if goPackage.UpdateVcs(); goPackage.Vcs != nil {

				goPackage.UpdateVcsFields()

				if goPackage.LocalBranch == goPackage.Vcs.GetDefaultBranch() &&
					goPackage.Status == "" &&
					goPackage.Local != goPackage.Remote {

					importPathElements := strings.Split(importPath, "/")
					if cc, _, err := gh.Repositories.CompareCommits(importPathElements[1], path.Join(importPathElements[2:]...), goPackage.Local, goPackage.Remote); err == nil {

						for _, repositoryCommit := range cc.Commits {
							if repositoryCommit.Commit != nil && repositoryCommit.Commit.Message != nil {
								fmt.Fprintln(w, *repositoryCommit.Commit.Message)
							}
						}
					}
				}
			}
		}
	}
}

func doStuffWithPackage(w io.Writer, goPackage *GoPackage) {
	if goPackage != nil {
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
			io.WriteString(w, "<p>"+presenter(goPackage)+"</p>")
		}
	} else {
		panic("Unexpected")
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

func mainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=us-ascii")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	fw := &FlushWriter{w: w, f: w.(http.Flusher)}

	io.WriteString(fw, `<html><head></head><body>`)

	doStuff(fw)

	/*for i := 0; i < 10; i++ {
		io.WriteString(fw, "<p>blah blah</p><br>")
		time.Sleep(time.Second)
	}*/

	io.WriteString(fw, "</body></html>")
}

func main() {
	http.HandleFunc("/all", mainHandler)
	http.HandleFunc("/", debugHandler)

	err := http.ListenAndServe(":8080", nil)
	CheckError(err)
}
