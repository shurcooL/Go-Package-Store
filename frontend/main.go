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
	"sync"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
	gpscomponent "github.com/shurcooL/Go-Package-Store/vcomponent"
	"github.com/shurcooL/go/gopherjs_http/jsutil"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func main() {
	js.Global.Set("UpdateRepository", jsutil.Wrap(UpdateRepository))
	js.Global.Set("UpdateAll", jsutil.Wrap(UpdateAll))

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
	err := stream()
	if err != nil {
		log.Println(err)
	}
}

func stream() error {
	// TODO: Initial render might not be needed if the server prerenders initial state.
	//err := renderBody()
	//if err != nil {
	//	return err
	//}
	vecty.RenderBody(body)

	resp, err := http.Get("/api/updates")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		var rp gpscomponent.RepoPresentation
		err := dec.Decode(&rp)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		rpsMu.Lock()
		rps = append(rps, &rp)
		moveUp(rps, &rp)
		rpsMu.Unlock()

		err = renderBody()
		if err != nil {
			return err
		}
		if len(rps) >= 10 {
			break
		}
	}
	checkingUpdates = false

	err = renderBody()
	return err
}

var (
	rpsMu sync.Mutex // TODO: Move towards a channel-based unified state manipulator.
	rps   []*gpscomponent.RepoPresentation

	checkingUpdates = true
)

func renderBody() error {
	started := time.Now()
	defer func() { fmt.Println("renderBody:", time.Since(started)) }()

	rpsMu.Lock()
	defer rpsMu.Unlock()

	//var buf bytes.Buffer
	//err := updates.RenderBodyInnerHTML(&buf, rps, checkingUpdates)
	//if err != nil {
	//	return err
	//}
	//
	//document.Body().SetInnerHTML(buf.String())
	//return nil

	vecty.Rerender(body)
	return nil
}

var body = &UpdatesBody{}

type UpdatesBody struct {
	vecty.Core

	//RPs             []*gpscomponent.RepoPresentation
	//CheckingUpdates bool
}

func (b *UpdatesBody) Render() *vecty.HTML {
	var ns = vecty.List{
		prop.Class("content"),

		&gpscomponent.UpdatesHeader{
			RPs:             rps,
			CheckingUpdates: checkingUpdates,
		},
	}
	for _, rp := range rps {
		ns = append(ns, &gpscomponent.RepoPresentation{
			RepoRoot:          rp.RepoRoot,
			ImportPathPattern: rp.ImportPathPattern,
			LocalRevision:     rp.LocalRevision,
			RemoteRevision:    rp.RemoteRevision,
			HomeURL:           rp.HomeURL,
			ImageURL:          rp.ImageURL,
			Changes:           rp.Changes,
			Error:             rp.Error,

			UpdateState: rp.UpdateState,

			// TODO: Find a place for this.
			UpdateSupported: rp.UpdateSupported,
		})
	}

	return elem.Body(
		&gpscomponent.Header{},
		elem.Div(
			prop.Class("center-max-width"),
			elem.Div(ns...),
		),
	)
}

// UpdateAll marks all available updates as updating, and performs updates in background in sequence.
func UpdateAll(event dom.Event) {
	event.PreventDefault()
	if event.(*dom.MouseEvent).Button != 0 {
		return
	}

	var updates []string // Repo roots to update.

	rpsMu.Lock()
	for _, rp := range rps {
		if rp.UpdateState == gpscomponent.Available {
			updates = append(updates, rp.RepoRoot)
			rp.UpdateState = gpscomponent.Updating
		}
	}
	rpsMu.Unlock()

	err := renderBody()
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for _, root := range updates {
			update(root)
		}
	}()
}

// UpdateRepository updates specified repository.
// root is the import path corresponding to the root of the repository.
func UpdateRepository(event dom.Event, root string) {
	event.PreventDefault()
	if event.(*dom.MouseEvent).Button != 0 {
		return
	}

	rpsMu.Lock()
	for _, rp := range rps {
		if rp.RepoRoot == root {
			rp.UpdateState = gpscomponent.Updating
			break
		}
	}
	rpsMu.Unlock()

	err := renderBody()
	if err != nil {
		log.Println(err)
		return
	}

	go update(root)
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

	rpsMu.Lock()
	moveDown(rps, root)
	for _, rp := range rps {
		if rp.RepoRoot == root {
			rp.UpdateState = gpscomponent.Updated
			break
		}
	}
	rpsMu.Unlock()

	err = renderBody()
	if err != nil {
		log.Println(err)
		return
	}
}

// moveDown moves root down the rps towards all other updated.
func moveDown(rps []*gpscomponent.RepoPresentation, root string) {
	var i int
	for ; rps[i].RepoRoot != root; i++ { // i is the current package about to be updated.
	}
	for ; i+1 < len(rps) && rps[i+1].UpdateState != gpscomponent.Updated; i++ {
		rps[i], rps[i+1] = rps[i+1], rps[i] // Swap the two.
	}
}

// moveUp moves last entry up the rps above all other updated entries, unless rp is already updated.
func moveUp(rps []*gpscomponent.RepoPresentation, rp *gpscomponent.RepoPresentation) {
	if rp.UpdateState == gpscomponent.Updated {
		return
	}
	for i := len(rps) - 1; i-1 >= 0 && rps[i-1].UpdateState == gpscomponent.Updated; i-- {
		rps[i], rps[i-1] = rps[i-1], rps[i] // Swap the two.
	}
}
