package component

import (
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
)

// UpdatesContent returns the entire content of updates tab.
func UpdatesContent(rps []*model.RepoPresentation, checkingUpdates bool) vecty.List {
	return vecty.List{
		&Header{},
		elem.Div(
			vecty.Markup(prop.Class("center-max-width")),
			elem.Div(
				updatesContent(rps, checkingUpdates),
			),
		),
	}
}

func updatesContent(rps []*model.RepoPresentation, checkingUpdates bool) vecty.List {
	var content = vecty.List{prop.Class("content")}

	content = append(content,
		updatesHeader{
			RPs:             rps,
			CheckingUpdates: checkingUpdates,
		}.Render()...,
	)

	wroteInstalledUpdates := false
	for _, rp := range rps {
		if rp.UpdateState == model.Updated && !wroteInstalledUpdates {
			content = append(content, InstalledUpdates())
			wroteInstalledUpdates = true
		}

		content = append(content, &RepoPresentation{
			RepoPresentation: rp,
		})
	}

	return content
}
