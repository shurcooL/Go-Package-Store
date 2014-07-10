package main

import (
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go/build"
	"go/token"

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
	return status.PorcelainPresenter(goPackage)[:3] == "  +" // Assumes status.PorcelainPresenter output is always at least 3 bytes.
}

func writeRepoCommonHat(w io.Writer, repo Repo) {
	goPackages := repo.goPackages

	var importPaths []string
	for _, goPackage := range goPackages {
		importPaths = append(importPaths, goPackage.Bpkg.ImportPath)
	}

	fmt.Fprintf(w, `<h3><span title="%s">%s <span class="smaller">(%d packages)</span></span></h3>`, strings.Join(importPaths, "\n"), repo.ImportPathPattern(), len(goPackages))
}

// TODO: Should really use html/template...
func GenerateGithubHtml(w io.Writer, repo Repo, cc *github.CommitsComparison) {
	//goon.DumpExpr(goPackage, cc)

	writeRepoCommonHat(w, repo)

	if cc.BaseCommit != nil && cc.BaseCommit.Author != nil && cc.BaseCommit.Author.AvatarURL != nil {
		// TODO: Factor out styles into css
		fmt.Fprintf(w, `<img style="float: left; border-radius: 4px;" src="%s" width="36" height="36">`, *cc.BaseCommit.Author.AvatarURL)
	}

	// TODO: Factor out styles into css
	fmt.Fprint(w, `<div style="float: right;">`)
	fmt.Fprintf(w, `<a href="javascript:void(0)" onclick="update_go_package(this);" id="%s" title="%s">Update</a>`, repo.ImportPathPattern(), fmt.Sprintf("go get -u -d %s", repo.ImportPathPattern()))
	fmt.Fprint(w, `</div>`)

	// TODO: Factor out styles into css
	// HACK: Manually aligned to the left of the image, this should be done via proper html layout
	fmt.Fprint(w, `<div style="padding-left: 36px;">`)
	fmt.Fprint(w, `<ul>`)

	for index := range cc.Commits {
		repositoryCommit := cc.Commits[len(cc.Commits)-1-index]
		if repositoryCommit.Commit != nil && repositoryCommit.Commit.Message != nil {
			fmt.Fprint(w, "<li>")
			fmt.Fprint(w, html.EscapeString(*repositoryCommit.Commit.Message))
			fmt.Fprint(w, "</li>")
		}
	}

	fmt.Fprint(w, `</ul>`)
	fmt.Fprint(w, `</div>`)
}

// TODO: Rename this to be more generic, since it will be used to handle all repos.
func GenerateGithubHtml2(w http.ResponseWriter, repo Repo, cc *github.CommitsComparison) {
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

func GenerateGenericHtml(w io.Writer, repo Repo) {
	writeRepoCommonHat(w, repo)

	// TODO: Factor out styles into css
	fmt.Fprint(w, `<div style="float: right;">`)
	fmt.Fprintf(w, `<a href="javascript:void(0)" onclick="update_go_package(this);" id="%s" title="%s">Update</a>`, repo.ImportPathPattern(), fmt.Sprintf("go get -u -d %s", repo.ImportPathPattern()))
	fmt.Fprint(w, `</div>`)

	fmt.Fprintf(w, `<div>unknown changes</div>`)
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
			if GetRepoImportPathPattern(goPackage.Dir.Repo.Vcs.RootPath(), goPackage.Bpkg.SrcRoot) == importPathPattern {
				fmt.Println("ExternallyUpdated", importPathPattern)
				ExternallyUpdated(goPackage.Dir.Repo.VcsLocal.GetSources()[1].(DepNode2ManualI))
				break
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
	//title := GetRepoImportPathPattern(repo.Vcs.RootPath(), goPackage.Bpkg.SrcRoot)
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

		if strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/") {
			// updateGithubHtml
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
				GenerateGithubHtml(w, repo, comparison.cc)
			}
		} else {
			updatesAvailable++
			GenerateGenericHtml(w, repo)
		}

		flusher.Flush()

		fmt.Printf("Part 2b: %v ms.\n", time.Since(started2).Seconds()*1000)
	}

	if updatesAvailable == 0 {
		io.WriteString(w, `<div><h2 style="text-align: center;">No Updates Available</h2></div>`)
	}

	fmt.Printf("Part 3: %v ms.\n", time.Since(started).Seconds()*1000)
}

func main2Handler(w http.ResponseWriter, r *http.Request) {
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

		if strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/") {
			// updateGithubHtml
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
				GenerateGithubHtml2(w, repo, comparison.cc)
			}
		} else {
			updatesAvailable++
			GenerateGithubHtml2(w, repo, nil)
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

type FlushWriter struct {
	w io.Writer
	f http.Flusher
}

func (fw *FlushWriter) Write(p []byte) (n int, err error) {
	defer fw.f.Flush()
	return fw.w.Write(p)
}

type data struct {
}

func (_ data) Names() chan string {
	names := make(chan string)
	go func() {
		for _, name := range []string{"Sunny", "Funny", "Joan", "Boohoo", "Mike"} {
			names <- name
			time.Sleep(time.Second)
		}
		close(names)
	}()
	return names
}

func (_ data) RepoCcs() chan RepoCc {
	ch := make(chan RepoCc)
	go func() {
		ch <- RepoCc{
			NewRepo("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml", []*GoPackage{goPackageXxx}),
			ccXxx,
		}
		close(ch)
	}()
	return ch
}

func devHandler(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles("./assets/dev.tmpl")
	if err != nil {
		log.Println("./assets/dev.tmpl:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := data{}
	err = t.Execute(&FlushWriter{w: w, f: w.(http.Flusher)}, data)
	if err != nil {
		log.Println("t.Execute:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type RepoCc struct {
	Repo Repo
	Cc   *github.CommitsComparison
}

func (repoCc RepoCc) AvatarUrl() template.URL {
	// THINK: Maybe use the repo owner avatar, instead of committer?
	if repoCc.Cc != nil && repoCc.Cc.BaseCommit != nil && repoCc.Cc.BaseCommit.Author != nil && repoCc.Cc.BaseCommit.Author.AvatarURL != nil {
		return template.URL(*repoCc.Cc.BaseCommit.Author.AvatarURL)
	}
	return "https://github.com/images/gravatars/gravatar-user-420.png"
}

func dev2Handler(w http.ResponseWriter, req *http.Request) {
	loadTemplates()

	CommonHat(w)
	fw := &FlushWriter{w: w, f: w.(http.Flusher)}
	defer CommonTail(fw)

	data := RepoCc{
		NewRepo("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml", []*GoPackage{goPackageXxx}),
		ccXxx,
	}
	err := t.Execute(fw, data)
	if err != nil {
		log.Println("t.Execute:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ---

var t *template.Template

func loadTemplates() {
	const filename = "./assets/dev2.tmpl"

	funcMap := template.FuncMap{
		"revIndex": func(index, length int) (revIndex int) { return (length - 1) - index },
	}

	var err error
	//t, err = template.ParseFiles(filename)
	t, err = template.New(filepath.Base(filename)).Funcs(funcMap).ParseFiles(filename)
	if err != nil {
		log.Println("loadTemplates: ./assets/dev2.tmpl:", err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	loadTemplates()

	goon.DumpExpr(os.Getwd())
	goon.DumpExpr(os.Getenv("PATH"), os.Getenv("GOPATH"))

	http.HandleFunc("/index", mainHandler)
	http.HandleFunc("/index2", main2Handler)
	http.HandleFunc("/-/update", updateHandler)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/assets/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/dev", devHandler)
	http.HandleFunc("/dev2", dev2Handler)

	u4.Open("http://localhost:7043/index2")
	//u4.Open("http://localhost:7043/dev")

	err := http.ListenAndServe("localhost:7043", nil)
	CheckError(err)
}

// TODO: Remove this dev stuff.

var goPackageXxx = (*GoPackage)(&GoPackage{
	Bpkg: (*build.Package)(&build.Package{
		Dir:         (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml"),
		Name:        (string)("toml"),
		Doc:         (string)("Package toml provides facilities for decoding TOML configuration files via reflection."),
		ImportPath:  (string)("github.com/BurntSushi/toml"),
		Root:        (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding"),
		SrcRoot:     (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src"),
		PkgRoot:     (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/pkg"),
		BinDir:      (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/bin"),
		Goroot:      (bool)(false),
		PkgObj:      (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/pkg/darwin_amd64/github.com/BurntSushi/toml.a"),
		AllTags:     ([]string)([]string{}),
		ConflictDir: (string)(""),
		GoFiles: ([]string)([]string{
			(string)("decode.go"),
			(string)("doc.go"),
			(string)("encode.go"),
			(string)("lex.go"),
			(string)("parse.go"),
			(string)("type_check.go"),
			(string)("type_fields.go"),
		}),
		CgoFiles:       ([]string)([]string{}),
		IgnoredGoFiles: ([]string)([]string{}),
		CFiles:         ([]string)([]string{}),
		CXXFiles:       ([]string)([]string{}),
		HFiles:         ([]string)([]string{}),
		SFiles:         ([]string)([]string{}),
		SwigFiles:      ([]string)([]string{}),
		SwigCXXFiles:   ([]string)([]string{}),
		SysoFiles:      ([]string)([]string{}),
		CgoCFLAGS:      ([]string)([]string{}),
		CgoCPPFLAGS:    ([]string)([]string{}),
		CgoCXXFLAGS:    ([]string)([]string{}),
		CgoLDFLAGS:     ([]string)([]string{}),
		CgoPkgConfig:   ([]string)([]string{}),
		Imports: ([]string)([]string{
			(string)("bufio"),
			(string)("encoding"),
			(string)("errors"),
			(string)("fmt"),
			(string)("io"),
			(string)("io/ioutil"),
			(string)("log"),
			(string)("reflect"),
			(string)("sort"),
			(string)("strconv"),
			(string)("strings"),
			(string)("sync"),
			(string)("time"),
			(string)("unicode/utf8"),
		}),
		ImportPos: (map[string][]token.Position)(map[string][]token.Position{
			(string)("reflect"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
					Offset:   (int)(62),
					Line:     (int)(8),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(700),
					Line:     (int)(22),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/type_fields.go"),
					Offset:   (int)(252),
					Line:     (int)(10),
					Column:   (int)(2),
				}),
			}),
			(string)("sort"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(711),
					Line:     (int)(23),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/type_fields.go"),
					Offset:   (int)(263),
					Line:     (int)(11),
					Column:   (int)(2),
				}),
			}),
			(string)("log"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
					Offset:   (int)(31),
					Line:     (int)(5),
					Column:   (int)(2),
				}),
			}),
			(string)("fmt"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
					Offset:   (int)(36),
					Line:     (int)(5),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(687),
					Line:     (int)(20),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/lex.go"),
					Offset:   (int)(24),
					Line:     (int)(4),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
					Offset:   (int)(24),
					Line:     (int)(4),
					Column:   (int)(2),
				}),
			}),
			(string)("io/ioutil"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
					Offset:   (int)(49),
					Line:     (int)(7),
					Column:   (int)(2),
				}),
			}),
			(string)("strings"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
					Offset:   (int)(73),
					Line:     (int)(9),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(730),
					Line:     (int)(25),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
					Offset:   (int)(49),
					Line:     (int)(7),
					Column:   (int)(2),
				}),
			}),
			(string)("bufio"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(656),
					Line:     (int)(17),
					Column:   (int)(2),
				}),
			}),
			(string)("encoding"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
					Offset:   (int)(24),
					Line:     (int)(4),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(665),
					Line:     (int)(18),
					Column:   (int)(2),
				}),
			}),
			(string)("time"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
					Offset:   (int)(84),
					Line:     (int)(10),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
					Offset:   (int)(60),
					Line:     (int)(8),
					Column:   (int)(2),
				}),
			}),
			(string)("errors"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(677),
					Line:     (int)(19),
					Column:   (int)(2),
				}),
			}),
			(string)("strconv"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(719),
					Line:     (int)(24),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
					Offset:   (int)(38),
					Line:     (int)(6),
					Column:   (int)(2),
				}),
			}),
			(string)("unicode/utf8"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/lex.go"),
					Offset:   (int)(31),
					Line:     (int)(5),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse.go"),
					Offset:   (int)(68),
					Line:     (int)(9),
					Column:   (int)(2),
				}),
			}),
			(string)("sync"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/type_fields.go"),
					Offset:   (int)(271),
					Line:     (int)(12),
					Column:   (int)(2),
				}),
			}),
			(string)("io"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode.go"),
					Offset:   (int)(43),
					Line:     (int)(6),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode.go"),
					Offset:   (int)(694),
					Line:     (int)(21),
					Column:   (int)(2),
				}),
			}),
		}),
		TestGoFiles: ([]string)([]string{
			(string)("decode_test.go"),
			(string)("encode_test.go"),
			(string)("lex_test.go"),
			(string)("out_test.go"),
			(string)("parse_test.go"),
		}),
		TestImports: ([]string)([]string{
			(string)("bytes"),
			(string)("encoding/json"),
			(string)("flag"),
			(string)("fmt"),
			(string)("log"),
			(string)("reflect"),
			(string)("strings"),
			(string)("testing"),
			(string)("time"),
		}),
		TestImportPos: (map[string][]token.Position)(map[string][]token.Position{
			(string)("fmt"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
					Offset:   (int)(41),
					Line:     (int)(5),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/out_test.go"),
					Offset:   (int)(32),
					Line:     (int)(5),
					Column:   (int)(2),
				}),
			}),
			(string)("log"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
					Offset:   (int)(48),
					Line:     (int)(6),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/lex_test.go"),
					Offset:   (int)(24),
					Line:     (int)(4),
					Column:   (int)(2),
				}),
			}),
			(string)("reflect"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
					Offset:   (int)(55),
					Line:     (int)(7),
					Column:   (int)(2),
				}),
			}),
			(string)("bytes"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode_test.go"),
					Offset:   (int)(24),
					Line:     (int)(4),
					Column:   (int)(2),
				}),
			}),
			(string)("encoding/json"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
					Offset:   (int)(24),
					Line:     (int)(4),
					Column:   (int)(2),
				}),
			}),
			(string)("testing"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
					Offset:   (int)(66),
					Line:     (int)(8),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/encode_test.go"),
					Offset:   (int)(33),
					Line:     (int)(5),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/lex_test.go"),
					Offset:   (int)(31),
					Line:     (int)(5),
					Column:   (int)(2),
				}),
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse_test.go"),
					Offset:   (int)(35),
					Line:     (int)(5),
					Column:   (int)(2),
				}),
			}),
			(string)("time"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/decode_test.go"),
					Offset:   (int)(77),
					Line:     (int)(9),
					Column:   (int)(2),
				}),
			}),
			(string)("flag"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/out_test.go"),
					Offset:   (int)(24),
					Line:     (int)(4),
					Column:   (int)(2),
				}),
			}),
			(string)("strings"): ([]token.Position)([]token.Position{
				(token.Position)(token.Position{
					Filename: (string)("/Users/Dmitri/Dropbox/Work/2013/GoLanding/src/github.com/BurntSushi/toml/parse_test.go"),
					Offset:   (int)(24),
					Line:     (int)(4),
					Column:   (int)(2),
				}),
			}),
		}),
		XTestGoFiles:   ([]string)([]string{}),
		XTestImports:   ([]string)([]string{}),
		XTestImportPos: (map[string][]token.Position)(map[string][]token.Position{}),
	}),
	Standard: (bool)(false),
})

var ccXxx = (*github.CommitsComparison)(&github.CommitsComparison{
	BaseCommit: (*github.RepositoryCommit)(&github.RepositoryCommit{
		SHA: (*string)(NewString("d7b4e27ae7df432264ca4ecf2dbec313ed01c330")),
		Commit: (*github.Commit)(&github.Commit{
			SHA: (*string)(nil),
			Author: (*github.CommitAuthor)(&github.CommitAuthor{
				Date:  (*time.Time)(nil),
				Name:  (*string)(NewString("Andrew Gallant")),
				Email: (*string)(NewString("jamslam@gmail.com")),
			}),
			Committer: (*github.CommitAuthor)(&github.CommitAuthor{
				Date:  (*time.Time)(nil),
				Name:  (*string)(NewString("Andrew Gallant")),
				Email: (*string)(NewString("jamslam@gmail.com")),
			}),
			Message: (*string)(NewString("Merge pull request #16 from nobonobo/master\n\nInfinite loop avoidance in Unexpected EOF")),
			Tree: (*github.Tree)(&github.Tree{
				SHA:     (*string)(NewString("7b938c31378d4b37c244f66d62400d8b3e44bfdd")),
				Entries: ([]github.TreeEntry)([]github.TreeEntry{}),
			}),
			Parents: ([]github.Commit)([]github.Commit{}),
			Stats:   (*github.CommitStats)(nil),
		}),
		Author: (*github.User)(&github.User{
			Login:       (*string)(NewString("BurntSushi")),
			ID:          (*int)(NewInt(456674)),
			URL:         (*string)(NewString("https://api.github.com/users/BurntSushi")),
			AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/c07104de771c3b6f6c30be8f592ef8f7?d=https%3A%2F%2Fidenticons.github.com%2Fa4f98968984cf211c9cdfdb95e1e4fbd.png&r=x")),
			GravatarID:  (*string)(NewString("c07104de771c3b6f6c30be8f592ef8f7")),
			Name:        (*string)(nil),
			Company:     (*string)(nil),
			Blog:        (*string)(nil),
			Location:    (*string)(nil),
			Email:       (*string)(nil),
			Hireable:    (*bool)(nil),
			PublicRepos: (*int)(nil),
			Followers:   (*int)(nil),
			Following:   (*int)(nil),
			CreatedAt:   (*github.Timestamp)(nil),
		}),
		Committer: (*github.User)(&github.User{
			Login:       (*string)(NewString("BurntSushi")),
			ID:          (*int)(NewInt(456674)),
			URL:         (*string)(NewString("https://api.github.com/users/BurntSushi")),
			AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/c07104de771c3b6f6c30be8f592ef8f7?d=https%3A%2F%2Fidenticons.github.com%2Fa4f98968984cf211c9cdfdb95e1e4fbd.png&r=x")),
			GravatarID:  (*string)(NewString("c07104de771c3b6f6c30be8f592ef8f7")),
			Name:        (*string)(nil),
			Company:     (*string)(nil),
			Blog:        (*string)(nil),
			Location:    (*string)(nil),
			Email:       (*string)(nil),
			Hireable:    (*bool)(nil),
			PublicRepos: (*int)(nil),
			Followers:   (*int)(nil),
			Following:   (*int)(nil),
			CreatedAt:   (*github.Timestamp)(nil),
		}),
		Parents: ([]github.Commit)([]github.Commit{
			(github.Commit)(github.Commit{
				SHA:       (*string)(NewString("2fffd0e6ca4b88558be4bcab497231c95270cd07")),
				Author:    (*github.CommitAuthor)(nil),
				Committer: (*github.CommitAuthor)(nil),
				Message:   (*string)(nil),
				Tree:      (*github.Tree)(nil),
				Parents:   ([]github.Commit)([]github.Commit{}),
				Stats:     (*github.CommitStats)(nil),
			}),
			(github.Commit)(github.Commit{
				SHA:       (*string)(NewString("ff98ae77642e0bf7f0e2b63857903f44d88f5b5e")),
				Author:    (*github.CommitAuthor)(nil),
				Committer: (*github.CommitAuthor)(nil),
				Message:   (*string)(nil),
				Tree:      (*github.Tree)(nil),
				Parents:   ([]github.Commit)([]github.Commit{}),
				Stats:     (*github.CommitStats)(nil),
			}),
		}),
		Message: (*string)(nil),
		Stats:   (*github.CommitStats)(nil),
		Files:   ([]github.CommitFile)([]github.CommitFile{}),
	}),
	Status:       (*string)(NewString("ahead")),
	AheadBy:      (*int)(NewInt(3)),
	BehindBy:     (*int)(NewInt(0)),
	TotalCommits: (*int)(NewInt(3)),
	Commits: ([]github.RepositoryCommit)([]github.RepositoryCommit{
		(github.RepositoryCommit)(github.RepositoryCommit{
			SHA: (*string)(NewString("629e931d4930dcd3dc393b700a6d4dcd487441b0")),
			Commit: (*github.Commit)(&github.Commit{
				SHA: (*string)(nil),
				Author: (*github.CommitAuthor)(&github.CommitAuthor{
					Date:  (*time.Time)(nil),
					Name:  (*string)(NewString("Rafal Jeczalik")),
					Email: (*string)(NewString("rjeczalik@gmail.com")),
				}),
				Committer: (*github.CommitAuthor)(&github.CommitAuthor{
					Date:  (*time.Time)(nil),
					Name:  (*string)(NewString("Rafal Jeczalik")),
					Email: (*string)(NewString("rjeczalik@gmail.com")),
				}),
				Message: (*string)(NewString("gofmt")),
				Tree: (*github.Tree)(&github.Tree{
					SHA:     (*string)(NewString("858831a3d12594b093954d3b27df62bd57e76b5f")),
					Entries: ([]github.TreeEntry)([]github.TreeEntry{}),
				}),
				Parents: ([]github.Commit)([]github.Commit{}),
				Stats:   (*github.CommitStats)(nil),
			}),
			Author: (*github.User)(&github.User{
				Login:       (*string)(NewString("rjeczalik")),
				ID:          (*int)(NewInt(1162017)),
				URL:         (*string)(NewString("https://api.github.com/users/rjeczalik")),
				AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/6d043eda71024cee583863d5619bdb6c?d=https%3A%2F%2Fidenticons.github.com%2F7c8268b4dc89926ce0772f124b811303.png&r=x")),
				GravatarID:  (*string)(NewString("6d043eda71024cee583863d5619bdb6c")),
				Name:        (*string)(nil),
				Company:     (*string)(nil),
				Blog:        (*string)(nil),
				Location:    (*string)(nil),
				Email:       (*string)(nil),
				Hireable:    (*bool)(nil),
				PublicRepos: (*int)(nil),
				Followers:   (*int)(nil),
				Following:   (*int)(nil),
				CreatedAt:   (*github.Timestamp)(nil),
			}),
			Committer: (*github.User)(&github.User{
				Login:       (*string)(NewString("rjeczalik")),
				ID:          (*int)(NewInt(1162017)),
				URL:         (*string)(NewString("https://api.github.com/users/rjeczalik")),
				AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/6d043eda71024cee583863d5619bdb6c?d=https%3A%2F%2Fidenticons.github.com%2F7c8268b4dc89926ce0772f124b811303.png&r=x")),
				GravatarID:  (*string)(NewString("6d043eda71024cee583863d5619bdb6c")),
				Name:        (*string)(nil),
				Company:     (*string)(nil),
				Blog:        (*string)(nil),
				Location:    (*string)(nil),
				Email:       (*string)(nil),
				Hireable:    (*bool)(nil),
				PublicRepos: (*int)(nil),
				Followers:   (*int)(nil),
				Following:   (*int)(nil),
				CreatedAt:   (*github.Timestamp)(nil),
			}),
			Parents: ([]github.Commit)([]github.Commit{
				(github.Commit)(github.Commit{
					SHA:       (*string)(NewString("d7b4e27ae7df432264ca4ecf2dbec313ed01c330")),
					Author:    (*github.CommitAuthor)(nil),
					Committer: (*github.CommitAuthor)(nil),
					Message:   (*string)(nil),
					Tree:      (*github.Tree)(nil),
					Parents:   ([]github.Commit)([]github.Commit{}),
					Stats:     (*github.CommitStats)(nil),
				}),
			}),
			Message: (*string)(nil),
			Stats:   (*github.CommitStats)(nil),
			Files:   ([]github.CommitFile)([]github.CommitFile{}),
		}),
		(github.RepositoryCommit)(github.RepositoryCommit{
			SHA: (*string)(NewString("6cab9f41ecc899af473584dbeff6e1814a098a6c")),
			Commit: (*github.Commit)(&github.Commit{
				SHA: (*string)(nil),
				Author: (*github.CommitAuthor)(&github.CommitAuthor{
					Date:  (*time.Time)(nil),
					Name:  (*string)(NewString("Rafal Jeczalik")),
					Email: (*string)(NewString("rjeczalik@gmail.com")),
				}),
				Committer: (*github.CommitAuthor)(&github.CommitAuthor{
					Date:  (*time.Time)(nil),
					Name:  (*string)(NewString("Rafal Jeczalik")),
					Email: (*string)(NewString("rjeczalik@gmail.com")),
				}),
				Message: (*string)(NewString("fix go vet warnings")),
				Tree: (*github.Tree)(&github.Tree{
					SHA:     (*string)(NewString("896b4c18dcc467cd0a58c3d3d71300849eea68b8")),
					Entries: ([]github.TreeEntry)([]github.TreeEntry{}),
				}),
				Parents: ([]github.Commit)([]github.Commit{}),
				Stats:   (*github.CommitStats)(nil),
			}),
			Author: (*github.User)(&github.User{
				Login:       (*string)(NewString("rjeczalik")),
				ID:          (*int)(NewInt(1162017)),
				URL:         (*string)(NewString("https://api.github.com/users/rjeczalik")),
				AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/6d043eda71024cee583863d5619bdb6c?d=https%3A%2F%2Fidenticons.github.com%2F7c8268b4dc89926ce0772f124b811303.png&r=x")),
				GravatarID:  (*string)(NewString("6d043eda71024cee583863d5619bdb6c")),
				Name:        (*string)(nil),
				Company:     (*string)(nil),
				Blog:        (*string)(nil),
				Location:    (*string)(nil),
				Email:       (*string)(nil),
				Hireable:    (*bool)(nil),
				PublicRepos: (*int)(nil),
				Followers:   (*int)(nil),
				Following:   (*int)(nil),
				CreatedAt:   (*github.Timestamp)(nil),
			}),
			Committer: (*github.User)(&github.User{
				Login:       (*string)(NewString("rjeczalik")),
				ID:          (*int)(NewInt(1162017)),
				URL:         (*string)(NewString("https://api.github.com/users/rjeczalik")),
				AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/6d043eda71024cee583863d5619bdb6c?d=https%3A%2F%2Fidenticons.github.com%2F7c8268b4dc89926ce0772f124b811303.png&r=x")),
				GravatarID:  (*string)(NewString("6d043eda71024cee583863d5619bdb6c")),
				Name:        (*string)(nil),
				Company:     (*string)(nil),
				Blog:        (*string)(nil),
				Location:    (*string)(nil),
				Email:       (*string)(nil),
				Hireable:    (*bool)(nil),
				PublicRepos: (*int)(nil),
				Followers:   (*int)(nil),
				Following:   (*int)(nil),
				CreatedAt:   (*github.Timestamp)(nil),
			}),
			Parents: ([]github.Commit)([]github.Commit{
				(github.Commit)(github.Commit{
					SHA:       (*string)(NewString("629e931d4930dcd3dc393b700a6d4dcd487441b0")),
					Author:    (*github.CommitAuthor)(nil),
					Committer: (*github.CommitAuthor)(nil),
					Message:   (*string)(nil),
					Tree:      (*github.Tree)(nil),
					Parents:   ([]github.Commit)([]github.Commit{}),
					Stats:     (*github.CommitStats)(nil),
				}),
			}),
			Message: (*string)(nil),
			Stats:   (*github.CommitStats)(nil),
			Files:   ([]github.CommitFile)([]github.CommitFile{}),
		}),
		(github.RepositoryCommit)(github.RepositoryCommit{
			SHA: (*string)(NewString("f8260fb5e94dba7ed68a2621b5c4fdc675bd3861")),
			Commit: (*github.Commit)(&github.Commit{
				SHA: (*string)(nil),
				Author: (*github.CommitAuthor)(&github.CommitAuthor{
					Date:  (*time.Time)(nil),
					Name:  (*string)(NewString("Andrew Gallant")),
					Email: (*string)(NewString("jamslam@gmail.com")),
				}),
				Committer: (*github.CommitAuthor)(&github.CommitAuthor{
					Date:  (*time.Time)(nil),
					Name:  (*string)(NewString("Andrew Gallant")),
					Email: (*string)(NewString("jamslam@gmail.com")),
				}),
				Message: (*string)(NewString("We want %s since errorf escapes some characters (like new lines), which turns them into strings.")),
				Tree: (*github.Tree)(&github.Tree{
					SHA:     (*string)(NewString("94a352d78ef7c5484d13f43663492e137988627b")),
					Entries: ([]github.TreeEntry)([]github.TreeEntry{}),
				}),
				Parents: ([]github.Commit)([]github.Commit{}),
				Stats:   (*github.CommitStats)(nil),
			}),
			Author: (*github.User)(&github.User{
				Login:       (*string)(NewString("BurntSushi")),
				ID:          (*int)(NewInt(456674)),
				URL:         (*string)(NewString("https://api.github.com/users/BurntSushi")),
				AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/c07104de771c3b6f6c30be8f592ef8f7?d=https%3A%2F%2Fidenticons.github.com%2Fa4f98968984cf211c9cdfdb95e1e4fbd.png&r=x")),
				GravatarID:  (*string)(NewString("c07104de771c3b6f6c30be8f592ef8f7")),
				Name:        (*string)(nil),
				Company:     (*string)(nil),
				Blog:        (*string)(nil),
				Location:    (*string)(nil),
				Email:       (*string)(nil),
				Hireable:    (*bool)(nil),
				PublicRepos: (*int)(nil),
				Followers:   (*int)(nil),
				Following:   (*int)(nil),
				CreatedAt:   (*github.Timestamp)(nil),
			}),
			Committer: (*github.User)(&github.User{
				Login:       (*string)(NewString("BurntSushi")),
				ID:          (*int)(NewInt(456674)),
				URL:         (*string)(NewString("https://api.github.com/users/BurntSushi")),
				AvatarURL:   (*string)(NewString("https://gravatar.com/avatar/c07104de771c3b6f6c30be8f592ef8f7?d=https%3A%2F%2Fidenticons.github.com%2Fa4f98968984cf211c9cdfdb95e1e4fbd.png&r=x")),
				GravatarID:  (*string)(NewString("c07104de771c3b6f6c30be8f592ef8f7")),
				Name:        (*string)(nil),
				Company:     (*string)(nil),
				Blog:        (*string)(nil),
				Location:    (*string)(nil),
				Email:       (*string)(nil),
				Hireable:    (*bool)(nil),
				PublicRepos: (*int)(nil),
				Followers:   (*int)(nil),
				Following:   (*int)(nil),
				CreatedAt:   (*github.Timestamp)(nil),
			}),
			Parents: ([]github.Commit)([]github.Commit{
				(github.Commit)(github.Commit{
					SHA:       (*string)(NewString("6cab9f41ecc899af473584dbeff6e1814a098a6c")),
					Author:    (*github.CommitAuthor)(nil),
					Committer: (*github.CommitAuthor)(nil),
					Message:   (*string)(nil),
					Tree:      (*github.Tree)(nil),
					Parents:   ([]github.Commit)([]github.Commit{}),
					Stats:     (*github.CommitStats)(nil),
				}),
			}),
			Message: (*string)(nil),
			Stats:   (*github.CommitStats)(nil),
			Files:   ([]github.CommitFile)([]github.CommitFile{}),
		}),
	}),
	Files: ([]github.CommitFile)([]github.CommitFile{
		(github.CommitFile)(github.CommitFile{
			SHA:       (*string)(NewString("27cea0fdf82c428519b9dbbd67df183853720c97")),
			Filename:  (*string)(NewString("encode_test.go")),
			Additions: (*int)(NewInt(14)),
			Deletions: (*int)(NewInt(14)),
			Changes:   (*int)(NewInt(28)),
			Status:    (*string)(NewString("modified")),
			Patch:     (*string)(NewString("@@ -75,29 +75,29 @@ func TestEncode(t *testing.T) {\n \t\t\t\tSliceOfMixedArrays    [][2]interface{}\n \t\t\t\tArrayOfMixedSlices    [2][]interface{}\n \t\t\t}{\n-\t\t\t\t[][2]int{[2]int{1, 2}, [2]int{3, 4}},\n-\t\t\t\t[2][]int{[]int{1, 2}, []int{3, 4}},\n+\t\t\t\t[][2]int{{1, 2}, {3, 4}},\n+\t\t\t\t[2][]int{{1, 2}, {3, 4}},\n \t\t\t\t[][2][]int{\n-\t\t\t\t\t[2][]int{\n-\t\t\t\t\t\t[]int{1, 2}, []int{3, 4},\n+\t\t\t\t\t{\n+\t\t\t\t\t\t{1, 2}, {3, 4},\n \t\t\t\t\t},\n-\t\t\t\t\t[2][]int{\n-\t\t\t\t\t\t[]int{5, 6}, []int{7, 8},\n+\t\t\t\t\t{\n+\t\t\t\t\t\t{5, 6}, {7, 8},\n \t\t\t\t\t},\n \t\t\t\t},\n \t\t\t\t[2][][2]int{\n-\t\t\t\t\t[][2]int{\n-\t\t\t\t\t\t[2]int{1, 2}, [2]int{3, 4},\n+\t\t\t\t\t{\n+\t\t\t\t\t\t{1, 2}, {3, 4},\n \t\t\t\t\t},\n-\t\t\t\t\t[][2]int{\n-\t\t\t\t\t\t[2]int{5, 6}, [2]int{7, 8},\n+\t\t\t\t\t{\n+\t\t\t\t\t\t{5, 6}, {7, 8},\n \t\t\t\t\t},\n \t\t\t\t},\n \t\t\t\t[][2]interface{}{\n-\t\t\t\t\t[2]interface{}{1, 2}, [2]interface{}{\"a\", \"b\"},\n+\t\t\t\t\t{1, 2}, {\"a\", \"b\"},\n \t\t\t\t},\n \t\t\t\t[2][]interface{}{\n-\t\t\t\t\t[]interface{}{1, 2}, []interface{}{\"a\", \"b\"},\n+\t\t\t\t\t{1, 2}, {\"a\", \"b\"},\n \t\t\t\t},\n \t\t\t},\n \t\t\twantOutput: `SliceOfArrays = [[1, 2], [3, 4]]\n@@ -162,8 +162,8 @@ ArrayOfMixedSlices = [[1, 2], [\"a\", \"b\"]]`,\n \t\t},\n \t\t\"nested map\": {\n \t\t\tinput: map[string]map[string]int{\n-\t\t\t\t\"a\": map[string]int{\"b\": 1},\n-\t\t\t\t\"c\": map[string]int{\"d\": 2},\n+\t\t\t\t\"a\": {\"b\": 1},\n+\t\t\t\t\"c\": {\"d\": 2},\n \t\t\t},\n \t\t\twantOutput: \"[a]\\n  b = 1\\n\\n[c]\\n  d = 2\",\n \t\t},")),
		}),
		(github.CommitFile)(github.CommitFile{
			SHA:       (*string)(NewString("43afe3c3fda0a46e22d0d66620f61c22e2e6a57e")),
			Filename:  (*string)(NewString("parse.go")),
			Additions: (*int)(NewInt(8)),
			Deletions: (*int)(NewInt(8)),
			Changes:   (*int)(NewInt(16)),
			Status:    (*string)(NewString("modified")),
			Patch:     (*string)(NewString("@@ -65,7 +65,7 @@ func parse(data string) (p *parser, err error) {\n \treturn p, nil\n }\n \n-func (p *parser) panic(format string, v ...interface{}) {\n+func (p *parser) panicf(format string, v ...interface{}) {\n \tmsg := fmt.Sprintf(\"Near line %d, key '%s': %s\",\n \t\tp.approxLine, p.current(), fmt.Sprintf(format, v...))\n \tpanic(parseError(msg))\n@@ -74,7 +74,7 @@ func (p *parser) panic(format string, v ...interface{}) {\n func (p *parser) next() item {\n \tit := p.lx.nextItem()\n \tif it.typ == itemError {\n-\t\tp.panic(\"Near line %d: %s\", it.line, it.val)\n+\t\tp.panicf(\"Near line %d: %s\", it.line, it.val)\n \t}\n \treturn it\n }\n@@ -164,7 +164,7 @@ func (p *parser) value(it item) (interface{}, tomlType) {\n \t\t\tif e, ok := err.(*strconv.NumError); ok &&\n \t\t\t\te.Err == strconv.ErrRange {\n \n-\t\t\t\tp.panic(\"Integer '%s' is out of the range of 64-bit \"+\n+\t\t\t\tp.panicf(\"Integer '%s' is out of the range of 64-bit \"+\n \t\t\t\t\t\"signed integers.\", it.val)\n \t\t\t} else {\n \t\t\t\tp.bug(\"Expected integer value, but got '%s'.\", it.val)\n@@ -184,7 +184,7 @@ func (p *parser) value(it item) (interface{}, tomlType) {\n \t\t\tif e, ok := err.(*strconv.NumError); ok &&\n \t\t\t\te.Err == strconv.ErrRange {\n \n-\t\t\t\tp.panic(\"Float '%s' is out of the range of 64-bit \"+\n+\t\t\t\tp.panicf(\"Float '%s' is out of the range of 64-bit \"+\n \t\t\t\t\t\"IEEE-754 floating-point numbers.\", it.val)\n \t\t\t} else {\n \t\t\t\tp.bug(\"Expected float value, but got '%s'.\", it.val)\n@@ -252,7 +252,7 @@ func (p *parser) establishContext(key Key, array bool) {\n \t\tcase map[string]interface{}:\n \t\t\thashContext = t\n \t\tdefault:\n-\t\t\tp.panic(\"Key '%s' was already created as a hash.\", keyContext)\n+\t\t\tp.panicf(\"Key '%s' was already created as a hash.\", keyContext)\n \t\t}\n \t}\n \n@@ -270,7 +270,7 @@ func (p *parser) establishContext(key Key, array bool) {\n \t\tif hash, ok := hashContext[k].([]map[string]interface{}); ok {\n \t\t\thashContext[k] = append(hash, make(map[string]interface{}))\n \t\t} else {\n-\t\t\tp.panic(\"Key '%s' was already created and cannot be used as \"+\n+\t\t\tp.panicf(\"Key '%s' was already created and cannot be used as \"+\n \t\t\t\t\"an array.\", keyContext)\n \t\t}\n \t} else {\n@@ -326,7 +326,7 @@ func (p *parser) setValue(key string, value interface{}) {\n \n \t\t// Otherwise, we have a concrete key trying to override a previous\n \t\t// key, which is *always* wrong.\n-\t\tp.panic(\"Key '%s' has already been defined.\", keyContext)\n+\t\tp.panicf(\"Key '%s' has already been defined.\", keyContext)\n \t}\n \thash[key] = value\n }\n@@ -411,7 +411,7 @@ func (p *parser) asciiEscapeToUnicode(s string) string {\n \t// UTF-8 characters like U+DCFF, but it doesn't.\n \tr := string(rune(hex))\n \tif !utf8.ValidString(r) {\n-\t\tp.panic(\"Escaped character '\\\\u%s' is not valid UTF-8.\", s)\n+\t\tp.panicf(\"Escaped character '\\\\u%s' is not valid UTF-8.\", s)\n \t}\n \treturn string(r)\n }")),
		}),
		(github.CommitFile)(github.CommitFile{
			SHA:       (*string)(NewString("b7897e79d2b19c3690d9e34e0c5fe7e71b5fd680")),
			Filename:  (*string)(NewString("toml-test-encoder/main.go")),
			Additions: (*int)(NewInt(2)),
			Deletions: (*int)(NewInt(2)),
			Changes:   (*int)(NewInt(4)),
			Status:    (*string)(NewString("modified")),
			Patch:     (*string)(NewString("@@ -59,8 +59,8 @@ func translate(typedJson interface{}) interface{} {\n \t\t\tif m, ok := translate(v[i]).(map[string]interface{}); ok {\n \t\t\t\ttabArray[i] = m\n \t\t\t} else {\n-\t\t\t\tlog.Fatalf(\"JSON arrays may only contain objects. This \"+\n-\t\t\t\t\t\"corresponds to only tables being allowed in \"+\n+\t\t\t\tlog.Fatalf(\"JSON arrays may only contain objects. This \" +\n+\t\t\t\t\t\"corresponds to only tables being allowed in \" +\n \t\t\t\t\t\"TOML table arrays.\")\n \t\t\t}\n \t\t}")),
		}),
		(github.CommitFile)(github.CommitFile{
			SHA:       (*string)(NewString("026ac6ae6d52914510b2a647f85ea513c3e822d9")),
			Filename:  (*string)(NewString("type_check.go")),
			Additions: (*int)(NewInt(1)),
			Deletions: (*int)(NewInt(1)),
			Changes:   (*int)(NewInt(2)),
			Status:    (*string)(NewString("modified")),
			Patch:     (*string)(NewString("@@ -70,7 +70,7 @@ func (p *parser) typeOfArray(types []tomlType) tomlType {\n \ttheType := types[0]\n \tfor _, t := range types[1:] {\n \t\tif !typeEqual(theType, t) {\n-\t\t\tp.panic(\"Array contains values of type '%s' and '%s', but arrays \"+\n+\t\t\tp.panicf(\"Array contains values of type '%s' and '%s', but arrays \"+\n \t\t\t\t\"must be homogeneous.\", theType, t)\n \t\t}\n \t}")),
		}),
	}),
})

func NewString(s string) *string {
	return &s
}

func NewInt(i int) *int {
	return &i
}
