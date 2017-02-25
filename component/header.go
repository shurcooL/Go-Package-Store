package component

import (
	"fmt"

	"github.com/shurcooL/htmlg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Header struct{}

func (Header) Render() []*html.Node {
	// TODO: Make this much nicer.
	/*
		<div style="width: 100%; text-align: center; background-color: hsl(209, 51%, 92%);">
			<span style="background-color: hsl(209, 51%, 88%); padding: 15px; display: inline-block;">Updates</span>
		</div>
	*/
	div := &html.Node{
		Type: html.ElementNode, Data: atom.Div.String(),
		Attr: []html.Attribute{
			{Key: atom.Style.String(), Val: "width: 100%; text-align: center; background-color: hsl(209, 51%, 92%);"},
		},
		FirstChild: &html.Node{
			Type: html.ElementNode, Data: atom.Span.String(),
			Attr: []html.Attribute{
				{Key: atom.Style.String(), Val: "background-color: hsl(209, 51%, 88%); padding: 15px; display: inline-block;"},
			},
			FirstChild: htmlg.Text("Updates"),
		},
	}
	return []*html.Node{div}
}

// UpdatesHeader combines checkingForUpdates, noUpdates and updatesHeading
// into one high level component.
type UpdatesHeader struct {
	RPs             []*RepoPresentation
	CheckingUpdates bool
}

func (u UpdatesHeader) Render() []*html.Node {
	var ns []*html.Node
	// Show "Checking for updates..." while still checking.
	if u.CheckingUpdates {
		ns = append(ns, checkingForUpdates.Render()...)
	}
	available, updating, supported := u.status()
	// Show "No Updates Available" if we're done checking and there are no remaining updates.
	if !u.CheckingUpdates && available == 0 && !updating {
		ns = append(ns, noUpdates.Render()...)
	}
	// Show number of updates available and Update All button.
	ns = append(ns, updatesHeading{
		Available:       available,
		Updating:        updating,
		UpdateSupported: supported, // TODO: Fetch this value from backend once.
	}.Render()...)
	return ns
}

// status returns available, updating, supported updates in u.RPs.
func (u UpdatesHeader) status() (available uint, updating bool, supported bool) {
	for _, rp := range u.RPs {
		switch rp.UpdateState {
		case Available:
			available++
			supported = rp.UpdateSupported
		case Updating:
			updating = true
		}
	}
	return available, updating, supported
}

// updatesHeading is a heading that displays number of updates available,
// whether updates are installing, and an Update All button.
type updatesHeading struct {
	Available uint
	Updating  bool

	// TODO: Find a place for this.
	UpdateSupported bool
}

func (u updatesHeading) Render() []*html.Node {
	if u.Available == 0 && !u.Updating {
		return nil
	}
	h4 := &html.Node{
		Type: html.ElementNode, Data: atom.H4.String(),
		Attr: []html.Attribute{
			{Key: atom.Style.String(), Val: "text-align: left;"},
		},
	}
	if u.Updating {
		h4.AppendChild(htmlg.Text("Updates Installing..."))
	} else {
		h4.AppendChild(htmlg.Text(fmt.Sprintf("%d Updates Available", u.Available)))
	}
	h4.AppendChild(&html.Node{
		Type: html.ElementNode, Data: atom.Span.String(),
		Attr: []html.Attribute{
			{Key: atom.Style.String(), Val: "float: right;"},
		},
		FirstChild: u.updateAllButton(),
	})
	return []*html.Node{h4}
}

func (u updatesHeading) updateAllButton() *html.Node {
	if !u.UpdateSupported {
		return &html.Node{
			Type: html.ElementNode, Data: atom.Span.String(),
			Attr: []html.Attribute{
				{Key: atom.Style.String(), Val: "color: gray; cursor: default;"},
				{Key: atom.Title.String(), Val: "Updating repos is not currently supported for this source of repos."},
			},
			FirstChild: htmlg.Text("Update All"),
		}
	}
	switch {
	case u.Available > 0:
		return &html.Node{
			Type: html.ElementNode, Data: atom.A.String(),
			Attr: []html.Attribute{
				{Key: atom.Href.String(), Val: "/api/update-all"}, // TODO: Should it be a separate endpoint or what?
				{Key: atom.Onclick.String(), Val: "UpdateAll(event);"},
			},
			FirstChild: htmlg.Text("Update All"),
		}
	case u.Available == 0:
		return &html.Node{
			Type: html.ElementNode, Data: atom.Span.String(),
			Attr: []html.Attribute{
				{Key: atom.Style.String(), Val: "color: gray; cursor: default;"},
			},
			FirstChild: htmlg.Text("Update All"),
		}
	default:
		panic("unreachable")
	}
}

// InstalledUpdates is a heading for installed updates.
var InstalledUpdates = heading{Heading: atom.H3, Text: "Installed Updates"}

var checkingForUpdates = heading{Heading: atom.H2, Text: "Checking for updates..."}

var noUpdates = heading{Heading: atom.H2, Text: "No Updates Available"}

type heading struct {
	Heading atom.Atom
	Text    string
}

func (h heading) Render() []*html.Node {
	// TODO: Make this much nicer.
	// <{{.Heading}} style="text-align: center;">{{.Text}}</{{.Heading}}>
	hn := &html.Node{
		Type: html.ElementNode, Data: h.Heading.String(),
		Attr: []html.Attribute{
			{Key: atom.Style.String(), Val: "text-align: center;"},
		},
		FirstChild: htmlg.Text(h.Text),
	}
	return []*html.Node{hn}
}
