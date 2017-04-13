package component

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
	"github.com/gopherjs/vecty/style"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
	"github.com/shurcooL/octiconssvg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// RepoPresentation is a component for presenting a repository update.
//
// TODO: Dedup with workspace.RepoPresentation. Maybe.
type RepoPresentation struct {
	vecty.Core
	*model.RepoPresentation
}

// Render renders the component.
func (p *RepoPresentation) Render() *vecty.HTML {
	return elem.Div(
		prop.Class("list-entry go-package-update"),
		vecty.Property(atom.Id.String(), p.RepoRoot),
		vecty.Style("position", "relative"),
		elem.Div(
			prop.Class("list-entry-header"),
			elem.Span(
				vecty.Property(atom.Title.String(), p.ImportPathPattern),
				p.importPathPattern(),
			),
			elem.Div(
				vecty.Style("float", "right"),
				p.updateState(),
			),
		),
		elem.Div(
			prop.Class("list-entry-body"),
			elem.Image(
				vecty.Style("float", "left"), vecty.Style("border-radius", string(style.Px(4))),
				vecty.Property(atom.Src.String(), p.ImageURL),
				vecty.Property(atom.Width.String(), "36"),
				vecty.Property(atom.Height.String(), "36"),
			),
			elem.Div(
				p.presentationChangesAndError()...,
			),
			elem.Div(
				vecty.Style("clear", "both"),
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
			style.Color("gray"), vecty.Style("cursor", "default"),
			vecty.Property(atom.Title.String(), "Updating repos is not currently supported for this source of repos."),
			vecty.Text("Update"),
		)
	}
	switch p.UpdateState {
	case model.Available:
		return elem.Anchor(
			prop.Href("/api/update"),
			event.Click(func(e *vecty.Event) {
				// TODO.
				fmt.Printf("UpdateRepository(%q)\n", p.RepoRoot)
				// TODO: Modifying underlying model is bad because Restore can't tell if something changed...
				p.UpdateState = model.Updating // TODO: Do this via action.
				started := time.Now()
				vecty.Rerender(p)
				fmt.Println("render RepoPresentation:", time.Since(started))
				js.Global.Get("UpdateRepository").Invoke(p.RepoRoot)

			}).PreventDefault(),
			vecty.Text("Update"),
		)
	case model.Updating:
		return elem.Span(
			style.Color("gray"), vecty.Style("cursor", "default"),
			vecty.Text("Updating..."),
		)
	case model.Updated:
		// TODO.
		return nil
	default:
		panic("unreachable")
	}
}

func (p *RepoPresentation) presentationChangesAndError() vecty.List {
	return vecty.List{
		&PresentationChanges{
			RepoPresentation: p.RepoPresentation,
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

// PresentationChanges is a component containing changes within an update.
type PresentationChanges struct {
	vecty.Core
	//Changes        []*Change
	//LocalRevision  string // Only needed if len(Changes) == 0.
	//RemoteRevision string // Only needed if len(Changes) == 0.
	*model.RepoPresentation // Only uses Changes, and if len(Changes) == 0, then LocalRevision and RemoteRevision.
}

// Restore is called when the component should restore itself against a
// previous instance of a component. The previous component may be nil or
// of a different type than this Restorer itself, thus a type assertion
// should be used.
//
// If skip = true is returned, restoration of this component's body is
// skipped. That is, the component is not rerendered. If the component can
// prove when Restore is called that the HTML rendered by Component.Render
// would not change, true should be returned.
//func (p *PresentationChanges) Restore(prev vecty.Component) (skip bool) {
//	//fmt.Print("Restore: ")
//	old, ok := prev.(*PresentationChanges)
//	if !ok {
//		//fmt.Println("not *PresentationChanges")
//		return false
//	}
//	_ = old //fmt.Println("old.RepoPresentation == p.RepoPresentation:", old.RepoPresentation == p.RepoPresentation)
//	return false
//	//return old.RepoPresentation == p.RepoPresentation
//}

// Render renders the component.
func (p *PresentationChanges) Render() *vecty.HTML {
	//fmt.Println("PresentationChanges.Render()")
	switch len(p.Changes) {
	default:
		ns := vecty.List{
			prop.Class("changes-list"),
		}
		//for _, c := range p.Changes {
		//	ns = append(ns, &Change{
		//		Change: c,
		//	})
		//}
		for i := range p.Changes { // TODO: Consider changing model.RepoPresentation.Changes type to []*Change to simplify this.
			ns = append(ns, &Change{
				Change: &p.Changes[i],
			})
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
	*model.Change
}

// Render renders the component.
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
				vecty.UnsafeHTML(octiconGitCommit),
			),
		),
		elem.Span(
			vecty.Style("float", "right"), vecty.Style("margin-right", string(style.Px(6))),
			&Comments{Comments: &c.Comments},
		),
	)
}

// Comments is a component for displaying a change discussion.
//
// TODO: Consider inlining this into Change component, we'll see.
type Comments struct {
	vecty.Core
	*model.Comments
}

// Render renders the component.
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
			style.Color("currentColor"), vecty.Style("margin-right", string(style.Px(4))),
			vecty.UnsafeHTML(octiconComment),
		),
		vecty.Text(fmt.Sprint(c.Count)),
	)
}

// CommitID is a component that displays a short commit ID, with the full one available in tooltip.
type CommitID struct {
	vecty.Core
	ID string
}

// Render renders the component.
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

var (
	octiconGitCommit = render(octiconssvg.GitCommit)
	octiconComment   = render(octiconssvg.Comment)
)

func render(icon func() *html.Node) string {
	var buf bytes.Buffer
	err := html.Render(&buf, icon())
	if err != nil {
		panic(err)
	}
	return buf.String()
}
