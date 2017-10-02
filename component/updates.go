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
				vecty.Markup(prop.Class("content")),
				&updatesHeader{RPs: rps, CheckingUpdates: checkingUpdates},
				updatesContent(rps),
			),
		),
	}
}

func updatesContent(rps []*model.RepoPresentation) vecty.List {
	var content vecty.List
	wroteInstalledUpdates := false
	for _, rp := range rps {
		if rp.UpdateState == model.Updated && !wroteInstalledUpdates {
			content = append(content,
				elem.Heading3(
					vecty.Markup(
						// This element is mixed with keyed siblings, we choose an
						// arbitrary key here that can not conflict with any siblings.
						vecty.Key("__gps_installed_updates"),
						vecty.Style("text-align", "center"),
					),
					vecty.Text("Installed Updates"),
				),
			)
			wroteInstalledUpdates = true
		}

		content = append(content, &RepoPresentation{
			RepoPresentation: rp,
		})
	}

	return content
}
