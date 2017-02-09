// Package component contains HTML components used by Go Package Store.
package component

import (
	"fmt"
	"strconv"

	"github.com/shurcooL/htmlg"
	"github.com/shurcooL/octiconssvg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// TODO: Dedup with workspace.RepoPresentation. Maybe.
type RepoPresentation struct {
	RepoRoot          string
	ImportPathPattern string
	LocalRevision     string
	RemoteRevision    string
	HomeURL           string
	ImageURL          string
	Changes           []Change
	Error             string

	Updated bool

	// TODO: Find a place for this.
	UpdateSupported bool
}

func (p RepoPresentation) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		<div class="list-entry go-package-update" id="{{.Repo.Root}}" style="position: relative;">
			<div class="list-entry-header">
				{{.importPathPattern()}}

				{{if (not .Updated)}}{{.updateButton()}}{{end}}
			</div>
			<div class="list-entry-body">
				<img style="float: left; border-radius: 4px;" src="{{.Presentation.Image}}" width="36" height="36">

				<div>
					{{presentationChangesAndError()}}
				</div>
				<div style="clear: both;"></div>
			</div>
		</div>
	*/
	innerDiv1 := htmlg.DivClass("list-entry-header",
		&html.Node{
			Type: html.ElementNode, Data: atom.Span.String(),
			Attr: []html.Attribute{
				{Key: atom.Title.String(), Val: p.ImportPathPattern},
			},
			FirstChild: p.importPathPattern(),
		},
	)
	if !p.Updated {
		innerDiv1.AppendChild(&html.Node{
			Type: html.ElementNode, Data: atom.Div.String(),
			Attr: []html.Attribute{
				{Key: atom.Style.String(), Val: "float: right;"},
			},
			FirstChild: p.updateButton(),
		})
	}
	innerDiv2 := htmlg.DivClass("list-entry-body",
		&html.Node{
			Type: html.ElementNode, Data: atom.Img.String(),
			Attr: []html.Attribute{
				{Key: atom.Style.String(), Val: "float: left; border-radius: 4px;"},
				{Key: atom.Src.String(), Val: p.ImageURL},
				{Key: atom.Width.String(), Val: "36"},
				{Key: atom.Height.String(), Val: "36"},
			},
		},
		htmlg.Div(
			p.presentationChangesAndError()...,
		),
		&html.Node{
			Type: html.ElementNode, Data: atom.Div.String(),
			Attr: []html.Attribute{{Key: atom.Style.String(), Val: "clear: both;"}},
		},
	)
	div := &html.Node{
		Type: html.ElementNode, Data: atom.Div.String(),
		Attr: []html.Attribute{
			{Key: atom.Class.String(), Val: "list-entry go-package-update"},
			{Key: atom.Id.String(), Val: p.RepoRoot},
			{Key: atom.Style.String(), Val: "position: relative;"},
		},
	}
	div.AppendChild(innerDiv1)
	div.AppendChild(innerDiv2)
	return []*html.Node{div}
}

func (p RepoPresentation) presentationChangesAndError() []*html.Node {
	/*
		{{render (presentationchanges .)}}
		{{with .Presentation.Error}}
			<p class="presentation-error"><strong>Error:</strong> {{.}}</p>
		{{end}}
	*/
	var ns []*html.Node
	ns = append(ns, PresentationChanges{
		Changes:        p.Changes,
		LocalRevision:  p.LocalRevision,
		RemoteRevision: p.RemoteRevision,
	}.Render()...)
	if p.Error != "" {
		n := &html.Node{
			Type: html.ElementNode, Data: atom.P.String(),
			Attr: []html.Attribute{{Key: atom.Class.String(), Val: "presentation-error"}},
		}
		n.AppendChild(htmlg.Strong("Error:"))
		n.AppendChild(htmlg.Text(" "))
		n.AppendChild(htmlg.Text(p.Error))
		ns = append(ns, n)
	}
	return ns
}

// TODO: Turn this into a maybeLink, etc.
func (p RepoPresentation) importPathPattern() *html.Node {
	/*
		<span title="{{.Repo.ImportPathPattern}}">
			{{if .Presentation.Home}}
				<a href="{{.Presentation.Home}}" target="_blank"><strong>{{.Repo.ImportPathPattern}}</strong></a>
			{{else}}
				<strong>{{.Repo.ImportPathPattern}}</strong>
			{{end}}
		</span>
	*/
	var importPathPattern *html.Node
	if p.HomeURL != "" {
		importPathPattern = &html.Node{
			Type: html.ElementNode, Data: atom.A.String(),
			Attr: []html.Attribute{
				{Key: atom.Href.String(), Val: p.HomeURL},
				// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
				{Key: atom.Target.String(), Val: "_blank"},
			},
			FirstChild: htmlg.Strong(p.ImportPathPattern),
		}
	} else {
		importPathPattern = htmlg.Strong(p.ImportPathPattern)
	}
	return importPathPattern
}

func (p RepoPresentation) updateButton() *html.Node {
	/*
		<div style="float: right;">
			{{if updateSupported}}
				<a href="/-/update" onclick="UpdateRepository(event, '{{.Repo.Root | json}}');" class="update-button" title="go get -u -d {{.Repo.ImportPathPattern}}">Update</a>
			{{else}}
				<span style="color: gray; cursor: default;" title="Updating repos is not currently supported for this source of repos.">Update</span>
			{{end}}
		</div>
	*/
	if p.UpdateSupported {
		return &html.Node{
			Type: html.ElementNode, Data: atom.A.String(),
			Attr: []html.Attribute{
				{Key: atom.Href.String(), Val: "/-/update"},
				{Key: atom.Onclick.String(), Val: fmt.Sprintf("UpdateRepository(event, %q);", strconv.Quote(p.RepoRoot))},
				{Key: atom.Class.String(), Val: "update-button"},
				{Key: atom.Title.String(), Val: fmt.Sprintf("go get -u -d %s", p.ImportPathPattern)},
			},
			FirstChild: htmlg.Text("Update"),
		}
	} else {
		return &html.Node{
			Type: html.ElementNode, Data: atom.Span.String(),
			Attr: []html.Attribute{
				{Key: atom.Style.String(), Val: "color: gray; cursor: default;"},
				{Key: atom.Title.String(), Val: "Updating repos is not currently supported for this source of repos."},
			},
			FirstChild: htmlg.Text("Update"),
		}
	}
}

type PresentationChanges struct {
	Changes        []Change
	LocalRevision  string // Only needed if len(Changes) == 0.
	RemoteRevision string // Only needed if len(Changes) == 0.
}

func (p PresentationChanges) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		{{with .Presentation.Changes}}
			<ul class="changes-list">
				{{range .}}{{render (change .)}}{{end}}
			</ul>
		{{else}}
			<div class="changes-list">
				unknown changes
				{{with .Repo.Local.Revision}}from {{render (commitID .)}}{{end}}
				{{with .Repo.Remote.Revision}}to {{render (commitID .)}}{{end}}
			</div>
		{{end}}
	*/
	switch len(p.Changes) {
	default:
		var ns []*html.Node
		for _, c := range p.Changes {
			ns = append(ns, c.Render()...)
		}
		ul := htmlg.ULClass("changes-list", ns...)
		return []*html.Node{ul}
	case 0:
		var ns []*html.Node
		ns = append(ns, htmlg.Text("unknown changes"))
		if p.LocalRevision != "" {
			ns = append(ns, htmlg.Text(" from "))
			ns = append(ns, CommitID{ID: p.LocalRevision}.Render()...)
		}
		if p.RemoteRevision != "" {
			ns = append(ns, htmlg.Text(" to "))
			ns = append(ns, CommitID{ID: p.RemoteRevision}.Render()...)
		}
		div := htmlg.DivClass("changes-list", ns...)
		return []*html.Node{div}
	}
}

// Change is a component for a single commit message.
type Change struct {
	Message  string   // Commit message of this change.
	URL      string   // URL of this change.
	Comments Comments // Comments on this change.
}

func (c Change) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		<li>
			{{.Message}}
			<span class="highlight-on-hover">
				<a href="{{.URL}}" target="_blank" style="color: gray;" title="Commit">
					<octiconssvg.GitCommit() />
				</a>
			</span>
			<span style="float: right; margin-right: 6px;">
				{{render (comments .Comments)}}
			</span>
		</li>
	*/
	span1 := htmlg.SpanClass("highlight-on-hover",
		&html.Node{
			Type: html.ElementNode, Data: atom.A.String(),
			Attr: []html.Attribute{
				{Key: atom.Href.String(), Val: c.URL},
				// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
				{Key: atom.Target.String(), Val: "_blank"},
				{Key: atom.Style.String(), Val: "color: gray;"},
				{Key: atom.Title.String(), Val: "Commit"},
			},
			FirstChild: octiconssvg.GitCommit(),
		},
	)
	span2 := &html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr: []html.Attribute{
			{Key: atom.Style.String(), Val: "float: right; margin-right: 6px;"},
		},
	}
	appendChildren(span2, c.Comments.Render()...)
	li := htmlg.LI(
		htmlg.Text(c.Message),
		span1,
		span2,
	)
	return []*html.Node{li}
}

// Comments is a component for displaying a change discussion.
// TODO: Consider inlining this into Change component, we'll see.
type Comments struct {
	Count int
	URL   string
}

func (c Comments) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		{{if .Count}}
		 	<a href="{{.URL}}" target="_blank" style="color: gray;" title="{{.Count}} comments"><octiconssvg.Comment() style="color: currentColor;" />{{.Count}}</a>
		{{end}}
	*/
	if c.Count == 0 {
		return nil
	}
	a := &html.Node{
		Type: html.ElementNode, Data: atom.A.String(),
		Attr: []html.Attribute{
			{Key: atom.Href.String(), Val: c.URL},
			// TODO: Add rel="noopener", see https://dev.to/ben/the-targetblank-vulnerability-by-example.
			{Key: atom.Target.String(), Val: "_blank"},
			{Key: atom.Style.String(), Val: "color: gray;"},
			{Key: atom.Title.String(), Val: fmt.Sprintf("%d comments", c.Count)},
		},
	}
	a.AppendChild(&html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr: []html.Attribute{
			{Key: atom.Style.String(), Val: "color: currentColor; margin-right: 4px;"},
		},
		FirstChild: octiconssvg.Comment(),
	})
	a.AppendChild(htmlg.Text(fmt.Sprint(c.Count)))
	return []*html.Node{a}
}

// CommitID is a component that displays a short commit ID, with the full one available in tooltip.
type CommitID struct {
	ID string
}

func (c CommitID) Render() []*html.Node {
	// TODO: Make this much nicer.
	// <abbr title="{{.}}"><code class="commitID">{{commitID .}}</code></abbr>{{end}}
	code := &html.Node{
		Type: html.ElementNode, Data: atom.Code.String(),
		Attr: []html.Attribute{
			{Key: atom.Class.String(), Val: "commitID"},
		},
		FirstChild: htmlg.Text(c.commitID()),
	}
	abbr := &html.Node{
		Type: html.ElementNode, Data: atom.Abbr.String(),
		Attr: []html.Attribute{
			{Key: atom.Title.String(), Val: c.ID},
		},
		FirstChild: code,
	}
	return []*html.Node{abbr}
}

func (c CommitID) commitID() string { return c.ID[:8] }

// appendChildren adds nodes cs as children of n.
func appendChildren(n *html.Node, cs ...*html.Node) {
	for _, c := range cs {
		n.AppendChild(c)
	}
}
