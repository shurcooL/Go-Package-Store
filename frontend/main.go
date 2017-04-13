// Command frontend runs on frontend of Go Package Store.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/shurcooL/Go-Package-Store/frontend/action"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
	"github.com/shurcooL/Go-Package-Store/frontend/store"
	gpscomponent "github.com/shurcooL/Go-Package-Store/vcomponent"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func main() {
	js.Global.Set("UpdateRepositoryV", UpdateRepositoryV)
	js.Global.Set("UpdateAllV", UpdateAllV)

	switch readyState := document.ReadyState(); readyState {
	case "loading":
		document.AddEventListener("DOMContentLoaded", false, func(dom.Event) {
			go run()
		})
	case "interactive", "complete":
		run()
	default:
		panic(fmt.Errorf("internal error: unexpected document.ReadyState value: %v", readyState))
	}
}

func run() {
	// Initial frontend render.
	vecty.RenderBody(body)

	go scheduler()

	err := stream()
	if err != nil {
		log.Println(err)
	}
}

func stream() error {
	started := time.Now()
	defer func() { fmt.Println("stream:", time.Since(started)) }()

	resp, err := http.Get("/api/updates")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for len(store.RPs()) < 10 {
		var rp model.RepoPresentation
		err := dec.Decode(&rp)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		apply(&action.AppendRP{RP: &rp})
		//renderBody()
	}

	apply(&action.DoneCheckingUpdates{})
	//renderBody()
	return nil
}

type actionAndResponse struct {
	Action action.Action
	RespCh chan<- action.Response
}

var (
	actionCh = make(chan actionAndResponse) // TODO: Consider/try buffered channel of size 10.
	renderCh <-chan time.Time
)

func apply(a action.Action) action.Response {
	respCh := make(chan action.Response)
	actionCh <- actionAndResponse{Action: a, RespCh: respCh}
	resp := <-respCh
	return resp
}

func scheduler() {
	//var renderOn = make(chan struct{})
	//close(renderOn)

	//forceRenderCh := time.Tick(5000 * time.Millisecond)

	for {
		select {
		case a := <-actionCh:
			resp := store.Apply(a.Action)
			a.RespCh <- resp

			// TODO: Don't render (needlessly) after *action.SetUpdating, etc.
			renderCh = time.After(10 * time.Millisecond)
		case <-renderCh:
			renderBody()
			renderCh = nil
			//runtime.Gosched()

			// TODO: Add another case that forces a re-render to happen at least once every
			//       500 milliseconds or so (in case there are never-ending actions that
			//       take a while to get through; we still want to display some progress).
			//case <-forceRenderCh:
			//	if renderCh == nil {
			//		break
			//	}
			//	renderBody()
			//	renderCh = nil
		}

		//time.Sleep(time.Second)
		//runtime.Gosched()
	}
}

func renderBody() {
	started := time.Now()
	defer func() { fmt.Println("renderBody:", time.Since(started)) }()

	vecty.Rerender(body)
}

var body = &UpdatesBody{}

type UpdatesBody struct {
	vecty.Core
}

func (b *UpdatesBody) Render() *vecty.HTML {
	return elem.Body(
		gpscomponent.UpdatesContent(
			store.RPs(),
			store.CheckingUpdates(),
		)...,
	)
}

// UpdateAllV marks all available updates as updating, and performs updates in background in sequence.
func UpdateAllV() {
	go func() {
		started := time.Now()
		defer func() { fmt.Println("update all:", time.Since(started)) }()

		resp := apply(&action.SetUpdatingAll{}).(*action.SetUpdatingAllResponse)
		//renderBody()

		for _, root := range resp.RepoRoots {
			update(root)
		}
	}()
}

// UpdateRepositoryV updates specified repository.
// root is the import path corresponding to the root of the repository.
func UpdateRepositoryV(root string) {
	go func() {
		apply(&action.SetUpdating{RepoRoot: root})
		// No need to render body because the component updated itself internally.
		// TODO: Improve and centralize this when-and-what-to-rerender logic, maybe?

		update(root)
	}()
}

// update updates specified repository.
// root is the import path corresponding to the root of the repository.
func update(root string) {
	resp, err := http.PostForm("/api/update", url.Values{"RepoRoot": {root}})
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	// TODO: Check response for success or not, etc.
	//       This is a great chance to display update errors in frontend!
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	apply(&action.SetUpdated{RepoRoot: root})
	//renderBody()
}
