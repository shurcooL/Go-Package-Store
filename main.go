package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/shurcooL/Go-Package-Store/presenters"
	"github.com/shurcooL/go/gists/gist7480523"
	"github.com/shurcooL/go/gists/gist7651991"
	"github.com/shurcooL/go/gists/gist7802150"

	//. "gist.github.com/7519227.git"
	"github.com/shurcooL/go-goon"
	"github.com/shurcooL/go/exp/14"
	"github.com/shurcooL/go/u/u4"
	"github.com/shurcooL/gostatus/status"
)

func CommonHat(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
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

func shouldPresentUpdate(goPackage *gist7480523.GoPackage) bool {
	return status.PlumbingPresenterV2(goPackage)[:3] == "  +" // Ignore stash.
}

func WriteRepoHtml(w http.ResponseWriter, repoPresenter presenter.Change) {
	err := t.Execute(w, repoPresenter)
	if err != nil {
		log.Println("t.Execute:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var goPackages exp14.GoPackageList = &exp14.GoPackages{SkipGoroot: true}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if *godepsFlag != "" {
			// TODO: Implement updating Godeps packages.
			log.Fatalln("updating Godeps packages isn't supported yet")
		}

		importPathPattern := r.PostFormValue("import_path_pattern")

		fmt.Println("go", "get", "-u", "-d", importPathPattern)

		cmd := exec.Command("go", "get", "-u", "-d", importPathPattern)

		out, err := cmd.CombinedOutput()
		fmt.Println("out:", string(out))
		goon.DumpExpr(err)

		gist7802150.MakeUpdated(goPackages)
		for _, goPackage := range goPackages.List() {
			if rootPath := getRootPath(goPackage); rootPath != "" {
				if gist7480523.GetRepoImportPathPattern(rootPath, goPackage.Bpkg.SrcRoot) == importPathPattern {
					fmt.Println("ExternallyUpdated", importPathPattern)
					gist7802150.ExternallyUpdated(goPackage.Dir.Repo.VcsLocal.GetSources()[1].(gist7802150.DepNode2ManualI))
					break
				}
			}
		}

		fmt.Println("done", importPathPattern)
	}
}

func getRootPath(goPackage *gist7480523.GoPackage) (rootPath string) {
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

func mainHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: When "finished", should not reload templates from disk on each request... Unless using a dev flag?
	if err := loadTemplates(); err != nil {
		fmt.Fprintln(w, "loadTemplates:", err)
		return
	}

	started := time.Now()

	CommonHat(w)
	defer CommonTail(w)

	io.WriteString(w, `<div id="checking_updates"><h2 style="text-align: center;">Checking for updates...</h2></div>`)
	defer io.WriteString(w, `<script>document.getElementById("checking_updates").style.display = "none";</script>`)

	flusher := w.(http.Flusher)
	flusher.Flush()

	fmt.Printf("Part 1: %v ms.\n", time.Since(started).Seconds()*1000)

	// rootPath -> []*gist7480523.GoPackage
	var goPackagesInRepo = make(map[string][]*gist7480523.GoPackage)

	// TODO: Use http.CloseNotifier, e.g. https://sourcegraph.com/github.com/donovanhide/eventsource/tree/master/server.go#L70

	gist7802150.MakeUpdated(goPackages)
	fmt.Printf("Part 1b: %v ms.\n", time.Since(started).Seconds()*1000)
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

	goon.DumpExpr(len(goPackages.List()))
	goon.DumpExpr(len(goPackagesInRepo))

	fmt.Printf("Part 2: %v ms.\n", time.Since(started).Seconds()*1000)

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
		return repo
	}
	outChan := gist7651991.GoReduce(inChan, 8, reduceFunc)

	for out := range outChan {
		started2 := time.Now()

		repo := out.(gist7480523.GoPackageRepo)

		/*goPackage := repo.GoPackages()[0]

		// TODO: Factor these out into a nice interface...
		var comparison *GithubComparison
		if strings.HasPrefix(goPackage.Bpkg.ImportPath, "github.com/") {
			var ok bool
			comparison, ok = githubComparisons[repo.rootPath]
			if !ok {
				comparison = NewGithubComparison(goPackage.Bpkg.ImportPath, goPackage.Dir.Repo.VcsLocal, goPackage.Dir.Repo.VcsRemote)
				githubComparisons[repo.rootPath] = comparison
			}
			gist7802150.MakeUpdated(comparison)

			if comparison.err != nil {
				fmt.Println("couldn't compare:", comparison.err)
			}
		} else if strings.HasPrefix(goPackage.Bpkg.ImportPath, "gopkg.in/") {
			// TODO: gopkg.in needs to be supported in a better, less duplicated, and ensured to be correct way.
			//       In fact, it's a good test point for support for generic change-description interface (i.e., for github repos, code.google.com, etc.).
			var ok bool
			comparison, ok = githubComparisons[repo.rootPath]
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
			gist7802150.MakeUpdated(comparison)

			if comparison.err != nil {
				fmt.Println("couldn't compare:", comparison.err)
			}
		} else if strings.HasPrefix(goPackage.Dir.Repo.VcsLocal.Remote, "https://github.com/") {
			var ok bool
			comparison, ok = githubComparisons[repo.rootPath]
			if !ok {
				afterPrefix := goPackage.Dir.Repo.VcsLocal.Remote[len("https://"):]
				importPath := strings.TrimSuffix(afterPrefix, ".git")
				comparison = NewGithubComparison(importPath, goPackage.Dir.Repo.VcsLocal, goPackage.Dir.Repo.VcsRemote)
				githubComparisons[repo.rootPath] = comparison
			}
			gist7802150.MakeUpdated(comparison)

			if comparison.err != nil {
				fmt.Println("couldn't compare:", comparison.err)
			}
		}*/

		repoPresenter := presenter.New(&repo)

		updatesAvailable++
		WriteRepoHtml(w, repoPresenter)

		flusher.Flush()

		fmt.Printf("Part 2b: %v ms.\n", time.Since(started2).Seconds()*1000)
	}

	if updatesAvailable == 0 {
		io.WriteString(w, `<div><h2 style="text-align: center;">No Updates Available</h2></div>`)
	}

	fmt.Printf("Part 3: %v ms.\n", time.Since(started).Seconds()*1000)
}

// ---

var t *template.Template

func loadTemplates() error {
	const filename = "./assets/repo.tmpl"

	var err error
	t, err = template.ParseFiles(filename)
	return err
}

var godepsFlag = flag.String("godeps", "", "Path to Godeps file to use.")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := loadTemplates()
	if err != nil {
		log.Fatalln("loadTemplates:", err)
	}

	flag.Parse()
	if *godepsFlag != "" {
		fmt.Println("Using Godeps file:", *godepsFlag)
		goPackages = NewGoPackagesFromGodeps(*godepsFlag)
	}

	goon.DumpExpr(os.Getwd())
	goon.DumpExpr(os.Getenv("PATH"), os.Getenv("GOPATH"))

	http.HandleFunc("/index", mainHandler)
	http.HandleFunc("/-/update", updateHandler)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/assets/", http.FileServer(http.Dir(".")))

	u4.Open("http://localhost:7043/index")

	err = http.ListenAndServe("localhost:7043", nil)
	if err != nil {
		panic(err)
	}
}
