// +build js

package component

import (
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
)

// UpdatesContent returns the entire content of updates tab.
func UpdatesContent(active, history []*model.RepoPresentation, checkingUpdates bool) []vecty.MarkupOrChild {
	return []vecty.MarkupOrChild{
		&Header{},
		elem.Div(
			vecty.Markup(vecty.Class("center-max-width")),
			elem.Div(
				updatesContent(active, history, checkingUpdates)...,
			),
		),
	}
}

func updatesContent(active, history []*model.RepoPresentation, checkingUpdates bool) []vecty.MarkupOrChild {
	var content = []vecty.MarkupOrChild{
		vecty.Markup(vecty.Class("content")),
	}

	// Updates header.
	content = append(content,
		updatesHeader{
			Active:          active,
			CheckingUpdates: checkingUpdates,
		}.Render()...,
	)

	// Active updates.
	for _, rp := range active {
		content = append(content, &RepoPresentation{
			RepoPresentation: rp,
		})
	}

	// History with "Recently Installed Updates" heading, if any.
	if len(history) > 0 {
		content = append(content, elem.Heading3(
			vecty.Markup(vecty.Style("text-align", "center"), vecty.Style("margin-top", "80px")),
			vecty.Text("Recently Installed Updates"),
		))

		for i := len(history) - 1; i >= 0; i-- {
			content = append(content, &RepoPresentation{
				RepoPresentation: history[i],
			})
		}
	}

	return content
}
