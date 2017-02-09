// Package component contains HTML components used by Go Package Store.
package component

import (
	"github.com/shurcooL/htmlg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type CommitID struct {
	ID string
}

func (c CommitID) Render() []*html.Node {
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
