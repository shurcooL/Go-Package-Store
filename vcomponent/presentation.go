package vcomponent

import (
	"bytes"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
	"github.com/shurcooL/octiconssvg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// TODO: Dedup with workspace.RepoPresentation. Maybe.
type RepoPresentation struct {
	vecty.Core

	RepoRoot          string
	ImportPathPattern string
	LocalRevision     string
	RemoteRevision    string
	HomeURL           string
	ImageURL          string
	Changes           []*Change
	Error             string

	UpdateState UpdateState

	// TODO: Find a place for this.
	UpdateSupported bool
}

type UpdateState uint8

const (
	Available UpdateState = iota
	Updating
	Updated
)

func (p *RepoPresentation) Render() *vecty.HTML {
	return elem.Div(
		prop.Class("list-entry go-package-update"),
		vecty.Property(atom.Id.String(), p.RepoRoot),
		vecty.Property(atom.Style.String(), "position: relative;"),
		elem.Div(
			prop.Class("list-entry-header"),
			elem.Span(
				vecty.Property(atom.Title.String(), p.ImportPathPattern),
				p.importPathPattern(),
			),
			elem.Div(
				vecty.Property(atom.Style.String(), "float: right;"),
				p.updateState(),
			),
		),
		elem.Div(
			prop.Class("list-entry-body"),
			elem.Image(
				vecty.Property(atom.Style.String(), "float: left; border-radius: 4px;"),
				vecty.Property(atom.Src.String(), p.ImageURL),
				vecty.Property(atom.Width.String(), "36"),
				vecty.Property(atom.Height.String(), "36"),
			),
			elem.Div(
				p.presentationChangesAndError()...,
			),
			elem.Div(
				vecty.Property(atom.Style.String(), "clear: both;"),
			),
		),
	)
}

// TODO: Turn this into a maybeLink, etc.
func (p *RepoPresentation) importPathPattern() *vecty.HTML {
	switch p.HomeURL {
	default:
		return elem.Anchor(
			prop.Href(p.HomeURL),
			// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
			vecty.Property(atom.Target.String(), "_blank"),
			elem.Strong(vecty.Text(p.ImportPathPattern)),
		)
	case "":
		return elem.Strong(vecty.Text(p.ImportPathPattern))
	}
}

func (p *RepoPresentation) updateState() *vecty.HTML {
	if !p.UpdateSupported {
		return elem.Span(
			vecty.Property(atom.Style.String(), "color: gray; cursor: default;"),
			vecty.Property(atom.Title.String(), "Updating repos is not currently supported for this source of repos."),
			vecty.Text("Update"),
		)
	}
	switch p.UpdateState {
	case Available:
		return elem.Anchor(
			prop.Href("/api/update"),
			event.Click(func(e *vecty.Event) {
				// TODO.
				fmt.Printf("UpdateRepositoryV(%q)\n", p.RepoRoot)
				p.UpdateState = Updating
				vecty.Rerender(p)
				js.Global.Get("UpdateRepositoryV").Invoke(p.RepoRoot)

			}).PreventDefault(),
			vecty.Text("Update"),
		)
	case Updating:
		return elem.Span(
			vecty.Property(atom.Style.String(), "color: gray; cursor: default;"),
			vecty.Text("Updating..."),
		)
	case Updated:
		// TODO.
		return nil
	default:
		panic("unreachable")
	}
}

func (p *RepoPresentation) presentationChangesAndError() []vecty.MarkupOrComponentOrHTML {
	var changes []*Change
	for _, c := range p.Changes {
		changes = append(changes, &Change{
			Message:  c.Message,
			URL:      c.URL,
			Comments: c.Comments,
		})
	}
	return vecty.List{
		&PresentationChanges{
			Changes:        changes,
			LocalRevision:  p.LocalRevision,
			RemoteRevision: p.RemoteRevision,
		},
		vecty.If(p.Error != "",
			elem.Paragraph(
				prop.Class("presentation-error"),
				elem.Strong(vecty.Text("Error:")),
				vecty.Text(" "),
				vecty.Text(p.Error),
			),
		),
	}
}

type PresentationChanges struct {
	vecty.Core
	Changes        []*Change
	LocalRevision  string // Only needed if len(Changes) == 0.
	RemoteRevision string // Only needed if len(Changes) == 0.
}

func (p *PresentationChanges) Render() *vecty.HTML {
	switch len(p.Changes) {
	default:
		ns := vecty.List{
			prop.Class("changes-list"),
		}
		//for i := range p.Changes { // TODO: Consider changing Changes type to []*Change to simplify this.
		//	ns = append(ns, &p.Changes[i])
		//}
		for _, c := range p.Changes {
			ns = append(ns, c)
		}
		return elem.UnorderedList(ns...)
	case 0:
		return elem.Div(
			prop.Class("changes-list"),
			vecty.Text("unknown changes"),
			vecty.If(p.LocalRevision != "",
				vecty.Text(" from "),
				&CommitID{ID: p.LocalRevision},
			),
			vecty.If(p.RemoteRevision != "",
				vecty.Text(" to "),
				&CommitID{ID: p.RemoteRevision},
			),
		)
	}
}

// Change is a component for a single commit message.
type Change struct {
	vecty.Core
	Message  string   // Commit message of this change.
	URL      string   // URL of this change.
	Comments Comments // Comments on this change.
}

func (c *Change) Render() *vecty.HTML {
	return elem.ListItem(
		vecty.Text(c.Message),
		elem.Span(
			prop.Class("highlight-on-hover"),
			elem.Anchor(
				prop.Href(c.URL),
				// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
				vecty.Property(atom.Target.String(), "_blank"),
				vecty.Style("color", "gray"),
				vecty.Property(atom.Title.String(), "Commit"),
				octicon(octiconssvg.GitCommit),
			),
		),
		elem.Span(
			vecty.Property(atom.Style.String(), "float: right; margin-right: 6px;"),
			&c.Comments,
		),
	)
}

// Comments is a component for displaying a change discussion.
// TODO: Consider inlining this into Change component, we'll see.
type Comments struct {
	vecty.Core
	Count int
	URL   string
}

func (c *Comments) Render() *vecty.HTML {
	if c.Count == 0 {
		return nil
	}
	return elem.Anchor(
		prop.Href(c.URL),
		// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
		vecty.Property(atom.Target.String(), "_blank"),
		vecty.Style("color", "gray"),
		vecty.Property(atom.Title.String(), fmt.Sprintf("%d comments", c.Count)),
		elem.Span(
			vecty.Property(atom.Style.String(), "color: currentColor; margin-right: 4px;"),
			octicon(octiconssvg.Comment),
		),
		vecty.Text(fmt.Sprint(c.Count)),
	)
}

// CommitID is a component that displays a short commit ID, with the full one available in tooltip.
type CommitID struct {
	vecty.Core
	ID string
}

func (c *CommitID) Render() *vecty.HTML {
	return elem.Abbreviation(
		vecty.Property(atom.Title.String(), c.ID),
		elem.Code(
			prop.Class("commitID"),
			vecty.Text(c.commitID()),
		),
	)
}

func (c *CommitID) commitID() string { return c.ID[:8] }

func octicon(icon func() *html.Node) vecty.Markup {
	var buf bytes.Buffer
	err := html.Render(&buf, icon())
	if err != nil {
		panic(err)
	}
	return vecty.UnsafeHTML(buf.String())
}
