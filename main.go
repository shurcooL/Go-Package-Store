package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
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
	"github.com/shurcooL/go/u/u4"
	"github.com/shurcooL/gostatus/status"
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
func CommonTail(w io.Writer) {
	io.WriteString(w, "</body></html>")
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
	return status.PlumbingPresenterV2(goPackage)[:3] == "  +" // Ignore stash.
}

func WriteRepoHtml(w http.ResponseWriter, repo Repo, cc *github.CommitsComparison) {
	data := RepoCc{
		repo,
		cc,
	}
	err := t.Execute(w, data)
	if err != nil {
		log.Println("t.Execute:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var goPackages = &exp14.GoPackages{SkipGoroot: true}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		importPathPattern := r.PostFormValue("import_path_pattern")

		fmt.Println("go", "get", "-u", "-d", importPathPattern)

		cmd := exec.Command("go", "get", "-u", "-d", importPathPattern)

		out, err := cmd.CombinedOutput()
		goon.DumpExpr(string(out), err)

		MakeUpdated(goPackages)
		for _, goPackage := range goPackages.Entries {
			if rootPath := getRootPath(goPackage); rootPath != "" {
				if GetRepoImportPathPattern(rootPath, goPackage.Bpkg.SrcRoot) == importPathPattern {
					fmt.Println("ExternallyUpdated", importPathPattern)
					ExternallyUpdated(goPackage.Dir.Repo.VcsLocal.GetSources()[1].(DepNode2ManualI))
					break
				}
			}
		}

		fmt.Println("done", importPathPattern)
	}
}

func getRootPath(goPackage *GoPackage) (rootPath string) {
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

type Repo struct {
	rootPath   string
	goPackages []*GoPackage
}

func NewRepo(rootPath string, goPackages []*GoPackage) Repo {
	return Repo{rootPath, goPackages}
}

func (repo Repo) ImportPathPattern() string {
	return GetRepoImportPathPattern(repo.rootPath, repo.goPackages[0].Bpkg.SrcRoot)
}

func (repo Repo) RootPath() string         { return repo.rootPath }
func (repo Repo) GoPackages() []*GoPackage { return repo.goPackages }

func (repo Repo) ImportPaths() string {
	var importPaths []string
	for _, goPackage := range repo.goPackages {
		importPaths = append(importPaths, goPackage.Bpkg.ImportPath)
	}
	return strings.Join(importPaths, "\n")
}

func (repo Repo) WebLink() *template.URL {
	goPackage := repo.goPackages[0]

	// TODO: Factor these out into a nice interface...
	switch {
	case strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/"):
		importPathElements := strings.Split(goPackage.Bpkg.ImportPath, "/")
		url := template.URL("https://github.com/" + importPathElements[1] + "/" + importPathElements[2])
		return &url
	case strings.HasPrefix(goPackage.Bpkg.ImportPath, "gopkg.in/"):
		// TODO
		return nil
	default:
		return nil
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	started := time.Now()

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

	MakeUpdated(goPackages)
	fmt.Printf("Part 1b: %v ms.\n", time.Since(started).Seconds()*1000)
	if false {
		for _, goPackage := range goPackages.Entries {
			if rootPath := getRootPath(goPackage); rootPath != "" {
				goPackagesInRepo[rootPath] = append(goPackagesInRepo[rootPath], goPackage)
			}
		}
	} else {
		inChan := make(chan interface{})
		go func() { // This needs to happen in the background because sending input will be blocked on reading output.
			for _, goPackage := range goPackages.Entries {
				inChan <- goPackage
			}
			close(inChan)
		}()
		reduceFunc := func(in interface{}) interface{} {
			goPackage := in.(*GoPackage)
			if rootPath := getRootPath(goPackage); rootPath != "" {
				return Repo{rootPath, []*GoPackage{goPackage}}
			}
			return nil
		}
		outChan := GoReduce(inChan, 64, reduceFunc)
		for out := range outChan {
			repo := out.(Repo)
			goPackagesInRepo[repo.rootPath] = append(goPackagesInRepo[repo.rootPath], repo.goPackages[0])
		}
	}

	goon.DumpExpr(len(goPackagesInRepo))

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
	go func() { // This needs to happen in the background because sending input will be blocked on reading output.
		for rootPath, goPackages := range goPackagesInRepo {
			inChan <- Repo{rootPath, goPackages}
		}
		close(inChan)
	}()
	outChan := GoReduce(inChan, 8, reduceFunc)

	for out := range outChan {
		started2 := time.Now()

		repo := out.(Repo)

		goPackage := repo.goPackages[0]

		// TODO: Factor these out into a nice interface...
		if strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/") {
			comparison, ok := githubComparisons[repo.rootPath]
			if !ok {
				comparison = NewGithubComparison(goPackage.Bpkg.ImportPath, goPackage.Dir.Repo.VcsLocal, goPackage.Dir.Repo.VcsRemote)
				githubComparisons[repo.rootPath] = comparison
			}
			MakeUpdated(comparison)

			if comparison.err != nil {
				fmt.Println("couldn't compare:", comparison.err)
			} else {
				updatesAvailable++
				WriteRepoHtml(w, repo, comparison.cc)
			}
		} else if strings.HasPrefix(goPackage.Bpkg.ImportPath, "gopkg.in/") {
			// TODO: gopkg.in needs to be supported in a better, less duplicated, and ensured to be correct way.
			//       In fact, it's a good test point for support for generic change-description interface (i.e., for github repos, code.google.com, etc.).
			comparison, ok := githubComparisons[repo.rootPath]
			if !ok {
				afterPrefix := goPackage.Bpkg.ImportPath[len("gopkg.in/"):]
				importPathElements0 := strings.Split(afterPrefix, ".")
				if len(importPathElements0) != 2 {
					log.Panicln("len(importPathElements0) != 2", importPathElements0)
				}
				importPathElements1 := strings.Split(importPathElements0[0], "/")
				importPath := "github.com/"
				if len(importPathElements1) == 1 { // gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
					importPath += "go-" + importPathElements1[0] + "/" + importPathElements1[0]
				} else if len(importPathElements1) == 2 { // gopkg.in/user/pkg.v3 -> github.com/user/pkg
					importPath += importPathElements1[0] + "/" + importPathElements1[1]
				} else {
					log.Panicln("len(importPathElements1) != 1 nor 2", importPathElements1)
				}
				comparison = NewGithubComparison(importPath, goPackage.Dir.Repo.VcsLocal, goPackage.Dir.Repo.VcsRemote)
				githubComparisons[repo.rootPath] = comparison
			}
			MakeUpdated(comparison)

			if comparison.err != nil {
				fmt.Println("couldn't compare:", comparison.err)
			} else {
				updatesAvailable++
				WriteRepoHtml(w, repo, comparison.cc)
			}
		} else {
			updatesAvailable++
			WriteRepoHtml(w, repo, nil)
		}

		flusher.Flush()

		fmt.Printf("Part 2b: %v ms.\n", time.Since(started2).Seconds()*1000)
	}

	if updatesAvailable == 0 {
		io.WriteString(w, `<div><h2 style="text-align: center;">No Updates Available</h2></div>`)
	}

	fmt.Printf("Part 3: %v ms.\n", time.Since(started).Seconds()*1000)
}

// ---

type RepoCc struct {
	Repo Repo
	Cc   *github.CommitsComparison
}

func (this RepoCc) AvatarUrl() template.URL {
	// THINK: Maybe use the repo owner avatar, instead of committer?
	if this.Cc != nil && this.Cc.BaseCommit != nil && this.Cc.BaseCommit.Author != nil && this.Cc.BaseCommit.Author.AvatarURL != nil {
		return template.URL(*this.Cc.BaseCommit.Author.AvatarURL)
	}
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}

// List of changes, starting with the most recent.
func (this RepoCc) Changes() <-chan github.RepositoryCommit {
	out := make(chan github.RepositoryCommit)
	go func() {
		for index := range this.Cc.Commits {
			out <- this.Cc.Commits[len(this.Cc.Commits)-1-index]
		}
		close(out)
	}()
	return out
}

// ---

var t *template.Template

func loadTemplates() {
	const filename = "./assets/repo.tmpl"

	var err error
	t, err = template.ParseFiles(filename)
	if err != nil {
		log.Println("loadTemplates:", filename, err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	loadTemplates()

	goon.DumpExpr(os.Getwd())
	goon.DumpExpr(os.Getenv("PATH"), os.Getenv("GOPATH"))

	http.HandleFunc("/index", mainHandler)
	http.HandleFunc("/-/update", updateHandler)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/assets/", http.FileServer(http.Dir(".")))

	u4.Open("http://localhost:7043/index")

	err := http.ListenAndServe("localhost:7043", nil)
	CheckError(err)
}
