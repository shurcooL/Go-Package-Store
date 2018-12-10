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
	gpscomponent "github.com/shurcooL/Go-Package-Store/component"
	"github.com/shurcooL/Go-Package-Store/frontend/action"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
	"github.com/shurcooL/Go-Package-Store/frontend/store"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func main() {
	js.Global.Set("UpdateRepository", UpdateRepository)
	js.Global.Set("UpdateAll", UpdateAll)

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

	// Start the scheduler loop.
	go scheduler()

	// Start streaming repo presentations from the backend.
	err := stream()
	if err != nil {
		log.Println(err)
	}
}

// stream streams the list of repo presentations from the backend,
// and appends them to the store as they arrive.
func stream() error {
	started := time.Now()
	defer func() { fmt.Println("stream:", time.Since(started)) }()

	resp, err := http.Get("/api/updates")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		var rp model.RepoPresentation
		err := dec.Decode(&rp)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		apply(&action.AppendRP{RP: &rp})
	}

	apply(&action.DoneCheckingUpdates{})
	return nil
}

// scheduler runs a loop that is responsible for
// applying actions to the store as they're made available,
// and rendering the body after processing new actions.
//
// It coalesces temporally adjacent actions, processing
// them in batches without performing rendering in between.
func scheduler() {
	var renderCh <-chan time.Time

	for {
		select {
		case a := <-actionCh:
			resp := store.Apply(a.Action)
			a.RespCh <- resp

			renderCh = time.After(10 * time.Millisecond)
		case <-renderCh:
			renderBody()
			renderCh = nil
		}
	}
}

// TODO: Consider using time.NewTimer and Timer.Stop instead of time.After.

var actionCh = make(chan actionAndResponse) // TODO: Consider/try buffered channel of size 10.

type actionAndResponse struct {
	Action action.Action
	RespCh chan<- action.Response
}

// apply applies the given action to the store,
// and returns the response.
func apply(a action.Action) action.Response {
	respCh := make(chan action.Response)
	actionCh <- actionAndResponse{Action: a, RespCh: respCh}
	resp := <-respCh
	return resp
}

func renderBody() {
	started := time.Now()
	defer func() { fmt.Println("renderBody:", time.Since(started)) }()

	vecty.Rerender(body)
}

var body = &UpdatesBody{}

// UpdatesBody is the entire body of the updates tab.
type UpdatesBody struct {
	vecty.Core
}

// Render renders the component.
func (b *UpdatesBody) Render() vecty.ComponentOrHTML {
	return elem.Body(
		gpscomponent.UpdatesContent(
			store.Active(),
			store.History(),
			store.CheckingUpdates(),
		)...,
	)
}

// UpdateAll marks all available updates as updating, and performs updates in background in sequence.
func UpdateAll() {
	go func() {
		started := time.Now()
		defer func() { fmt.Println("update all:", time.Since(started)) }()

		resp := apply(&action.SetUpdatingAll{}).(*action.SetUpdatingAllResponse)

		for _, root := range resp.RepoRoots {
			update(root)
		}
	}()
}

// UpdateRepository updates specified repository.
// root is the import path corresponding to the root of the repository.
func UpdateRepository(root string) {
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
	started := time.Now()
	defer func() { fmt.Println("update:", time.Since(started)) }()

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
}
