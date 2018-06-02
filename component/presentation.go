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
	"github.com/shurcooL/octicon"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// RepoPresentation is a component for presenting a repository update.
//
// TODO: Dedup with workspace.RepoPresentation. Maybe.
type RepoPresentation struct {
	vecty.Core
	*model.RepoPresentation `vecty:"prop"`
}

// Render renders the component.
func (p *RepoPresentation) Render() vecty.ComponentOrHTML {
	return elem.Div(
		vecty.Markup(
			vecty.Class("list-entry", "go-package-update"),
			vecty.Property(atom.Id.String(), p.RepoRoot),
			vecty.Style("position", "relative"),
		),
		elem.Div(
			vecty.Markup(vecty.Class("list-entry-header")),
			elem.Span(
				vecty.Markup(vecty.Property(atom.Title.String(), p.ImportPathPattern)),
				p.importPathPattern(),
			),
			elem.Div(
				vecty.Markup(vecty.Style("float", "right")),
				p.updateState(),
			),
		),
		elem.Div(
			vecty.Markup(vecty.Class("list-entry-body")),
			elem.Image(
				vecty.Markup(
					vecty.Style("float", "left"), vecty.Style("border-radius", string(style.Px(4))),
					vecty.Property(atom.Src.String(), p.ImageURL),
					vecty.Property(atom.Width.String(), "36"),
					vecty.Property(atom.Height.String(), "36"),
				),
			),
			elem.Div(
				p.presentationChangesAndError()...,
			),
			elem.Div(
				vecty.Markup(vecty.Style("clear", "both")),
			),
		),
	)
}

// TODO: Turn this into a maybeLink, etc.
func (p *RepoPresentation) importPathPattern() *vecty.HTML {
	switch p.HomeURL {
	default:
		return elem.Anchor(
			vecty.Markup(
				prop.Href(p.HomeURL),
				// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
				vecty.Property(atom.Target.String(), "_blank"),
			),
			elem.Strong(vecty.Text(p.ImportPathPattern)),
		)
	case "":
		return elem.Strong(vecty.Text(p.ImportPathPattern))
	}
}

func (p *RepoPresentation) updateState() *vecty.HTML {
	if !p.UpdateSupported {
		return elem.Span(
			vecty.Markup(
				style.Color("gray"), vecty.Style("cursor", "default"),
				vecty.Property(atom.Title.String(), "Updating repos is not currently supported for this source of repos."),
			),
			vecty.Text("Update"),
		)
	}
	switch p.UpdateState {
	case model.Available:
		return elem.Anchor(
			vecty.Markup(
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
			),
			vecty.Text("Update"),
		)
	case model.Updating:
		return elem.Span(
			vecty.Markup(style.Color("gray"), vecty.Style("cursor", "default")),
			vecty.Text("Updating..."),
		)
	case model.Updated:
		// TODO.
		return nil
	default:
		panic("unreachable")
	}
}

func (p *RepoPresentation) presentationChangesAndError() []vecty.MarkupOrChild {
	return []vecty.MarkupOrChild{
		vecty.Markup(vecty.Style("word-break", "break-word")),
		&PresentationChanges{
			RepoPresentation: p.RepoPresentation,
		},
		vecty.If(p.Error != "",
			elem.Paragraph(
				vecty.Markup(vecty.Class("presentation-error")),
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
	*model.RepoPresentation `vecty:"prop"` // Only uses Changes, and if len(Changes) == 0, then LocalRevision and RemoteRevision.
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
func (p *PresentationChanges) Render() vecty.ComponentOrHTML {
	//fmt.Println("PresentationChanges.Render()")
	switch len(p.Changes) {
	default:
		ns := []vecty.MarkupOrChild{
			vecty.Markup(vecty.Class("changes-list")),
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
			vecty.Markup(vecty.Class("changes-list")),
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
	*model.Change `vecty:"prop"`
}

// Render renders the component.
func (c *Change) Render() vecty.ComponentOrHTML {
	return elem.ListItem(
		vecty.Text(c.Message),
		elem.Span(
			vecty.Markup(vecty.Class("highlight-on-hover")),
			elem.Anchor(
				vecty.Markup(
					prop.Href(c.URL),
					// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
					vecty.Property(atom.Target.String(), "_blank"),
					vecty.Style("color", "gray"),
					vecty.Property(atom.Title.String(), "Commit"),
					vecty.UnsafeHTML(octiconGitCommit),
				),
			),
		),
		elem.Span(
			vecty.Markup(vecty.Style("float", "right"), vecty.Style("margin-right", string(style.Px(6)))),
			&Comments{Comments: &c.Comments},
		),
	)
}

// Comments is a component for displaying a change discussion.
//
// TODO: Consider inlining this into Change component, we'll see.
type Comments struct {
	vecty.Core
	*model.Comments `vecty:"prop"`
}

// Render renders the component.
func (c *Comments) Render() vecty.ComponentOrHTML {
	if c.Count == 0 {
		return nil
	}
	return elem.Anchor(
		vecty.Markup(
			prop.Href(c.URL),
			// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
			vecty.Property(atom.Target.String(), "_blank"),
			vecty.Style("color", "gray"),
			vecty.Property(atom.Title.String(), fmt.Sprintf("%d comments", c.Count)),
		),
		elem.Span(
			vecty.Markup(
				style.Color("currentColor"), vecty.Style("margin-right", string(style.Px(4))),
				vecty.UnsafeHTML(octiconComment),
			),
		),
		vecty.Text(fmt.Sprint(c.Count)),
	)
}

// CommitID is a component that displays a short commit ID, with the full one available in tooltip.
type CommitID struct {
	vecty.Core
	ID string `vecty:"prop"`
}

// Render renders the component.
func (c *CommitID) Render() vecty.ComponentOrHTML {
	return elem.Abbreviation(
		vecty.Markup(vecty.Property(atom.Title.String(), c.ID)),
		elem.Code(
			vecty.Markup(vecty.Class("commitID")),
			vecty.Text(c.commitID()),
		),
	)
}

func (c *CommitID) commitID() string { return c.ID[:8] }

var (
	octiconGitCommit = render(octicon.GitCommit)
	octiconComment   = render(octicon.Comment)
)

func render(icon func() *html.Node) string {
	var buf bytes.Buffer
	err := html.Render(&buf, icon())
	if err != nil {
		panic(err)
	}
	return buf.String()
}
