// +build js

package main

import (
	"net/url"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/go/gopherjs_http/jsutil"
	"honnef.co/go/js/dom"
	"honnef.co/go/js/xhr"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func main() {
	js.Global.Set("UpdateRepository", jsutil.Wrap(UpdateRepository))
}

// UpdateRepository updates specified repository.
// repoRoot is the import path corresponding to the root of the repository.
func UpdateRepository(event dom.Event, repoRoot string) {
	event.PreventDefault()
	if event.(*dom.MouseEvent).Button != 0 {
		return
	}

	repoUpdate := document.GetElementByID(repoRoot)
	updateButton := repoUpdate.GetElementsByClassName("update-button")[0].(*dom.HTMLAnchorElement)

	updateButton.SetTextContent("Updating...")
	updateButton.AddEventListener("click", false, func(event dom.Event) { event.PreventDefault() })
	updateButton.SetTabIndex(-1)
	updateButton.Class().Add("disabled")

	go func() {
		req := xhr.NewRequest("POST", "/-/update")
		req.SetRequestHeader("Content-Type", "application/x-www-form-urlencoded")
		err := req.Send(url.Values{"repo_root": {repoRoot}}.Encode())
		if err != nil {
			println(err.Error())
			return
		}

		// Hide the "Updating..." label.
		updateButton.Style().SetProperty("display", "none", "")

		// Show "No Updates Available" if there are no remaining updates.
		if !anyUpdatesRemaining() {
			document.GetElementByID("no_updates").(dom.HTMLElement).Style().SetProperty("display", "", "")
		}

		// Move this Go package to "Installed Updates" list.
		installedUpdates := document.GetElementByID("installed_updates").(dom.HTMLElement)
		installedUpdates.Style().SetProperty("display", "", "")
		installedUpdates.ParentNode().InsertBefore(repoUpdate, installedUpdates.NextSibling()) // Insert after.
	}()
}

// anyUpdatesRemaining reports if there's at least one available or in-flight update.
func anyUpdatesRemaining() bool {
	updates := document.GetElementsByClassName("go-package-update")
	for _, update := range updates {
		updateButton := update.GetElementsByClassName("update-button")[0].(*dom.HTMLAnchorElement)
		updateButtonVisible := updateButton.Style().GetPropertyValue("display") != "none"
		if updateButtonVisible {
			return true
		}
	}
	return false
}
