// Command frontend runs on frontend of Go Package Store.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	gpscomponent "github.com/shurcooL/Go-Package-Store/component"
	"github.com/shurcooL/go/gopherjs_http/jsutil"
	"github.com/shurcooL/htmlg"
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
	err := renderBody()
	if err != nil {
		return err
	}

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
	rpsMu.Lock()
	defer rpsMu.Unlock()

	var buf bytes.Buffer

	err := htmlg.RenderComponents(&buf, gpscomponent.Header{})
	if err != nil {
		return err
	}

	_, err = io.WriteString(&buf, `<div class="center-max-width"><div class="content">`)
	if err != nil {
		return err
	}

	err = htmlg.RenderComponents(&buf, gpscomponent.UpdatesHeader{
		RPs:             rps,
		CheckingUpdates: checkingUpdates,
	})
	if err != nil {
		return err
	}

	wroteInstalledUpdates := false
	for _, rp := range rps {
		if rp.UpdateState == gpscomponent.Updated && !wroteInstalledUpdates {
			err = htmlg.RenderComponents(&buf, gpscomponent.InstalledUpdates)
			if err != nil {
				return err
			}
			wroteInstalledUpdates = true
		}

		err := htmlg.RenderComponents(&buf, rp)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(&buf, `</div></div>`)
	if err != nil {
		return err
	}

	document.Body().SetInnerHTML(buf.String())
	return nil
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
