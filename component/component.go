// Package component contains HTML components used by Go Package Store.
package component

import (
	"fmt"

	"github.com/shurcooL/htmlg"
	"github.com/shurcooL/octiconssvg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
