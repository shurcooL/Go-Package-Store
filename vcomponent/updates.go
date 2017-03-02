package vcomponent

import (
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

func UpdatesContent(rps []*RepoPresentation, checkingUpdates bool) vecty.List {
	return vecty.List{
		&Header{},
		elem.Div(
			prop.Class("center-max-width"),
			elem.Div(
				updatesContent(rps, checkingUpdates)...,
			),
		),
	}
}

func updatesContent(rps []*RepoPresentation, checkingUpdates bool) vecty.List {
	var content = vecty.List{prop.Class("content")}

	content = append(content,
		updatesHeader{
			RPs:             rps,
			CheckingUpdates: checkingUpdates,
		}.Render()...,
	)

	wroteInstalledUpdates := false
	for _, rp := range rps {
		if rp.UpdateState == Updated && !wroteInstalledUpdates {
			content = append(content, InstalledUpdates())
			wroteInstalledUpdates = true
		}

		content = append(content, &RepoPresentation{
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

	return content
}
