// Package component contains HTML components used by Go Package Store.
package component

import (
	"fmt"

	"github.com/shurcooL/htmlg"
	"github.com/shurcooL/octiconssvg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
