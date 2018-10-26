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

	// History with "Recently Installed Updates" heading, if any.
	if len(history) > 0 {
		content = append(content, heading(elem.Heading3, "Recently Installed Updates"))

		for _, rp := range history {
			content = append(content, &RepoPresentation{
				RepoPresentation: rp,
			})
		}

		// Spacer at the bottom.
		content = append(content, elem.Div(
			vecty.Markup(vecty.Style("height", "60px")),
		))
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

	return content
}
