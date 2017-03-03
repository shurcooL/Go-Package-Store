package vcomponent

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
	"golang.org/x/net/html/atom"
)

type Header struct {
	vecty.Core
}

func (*Header) Render() *vecty.HTML {
	return elem.Div(
		vecty.Property(atom.Style.String(), "width: 100%; text-align: center; background-color: hsl(209, 51%, 92%);"),
		elem.Span(
			vecty.Property(atom.Style.String(), "background-color: hsl(209, 51%, 88%); padding: 15px; display: inline-block;"),
			vecty.Text("Updates"),
		),
	)
}

// updatesHeader combines checkingForUpdates, noUpdates and updatesHeading
// into one high level component.
type updatesHeader struct {
	RPs             []*model.RepoPresentation
	CheckingUpdates bool
}

func (u updatesHeader) Render() vecty.List {
	var ns vecty.List
	// Show "Checking for updates..." while still checking.
	if u.CheckingUpdates {
		ns = append(ns, checkingForUpdates())
	}
	available, updating, supported := u.status()
	// Show "No Updates Available" if we're done checking and there are no remaining updates.
	if !u.CheckingUpdates && available == 0 && !updating {
		ns = append(ns, noUpdates())
	}
	// Show number of updates available and Update All button.
	ns = append(ns, &updatesHeading{
		Available:       available,
		Updating:        updating,
		UpdateSupported: supported, // TODO: Fetch this value from backend once.
	})
	return ns
}

// status returns available, updating, supported updates in u.RPs.
func (u updatesHeader) status() (available uint, updating bool, supported bool) {
	for _, rp := range u.RPs {
		switch rp.UpdateState {
		case model.Available:
			available++
			supported = rp.UpdateSupported
		case model.Updating:
			updating = true
		}
	}
	return available, updating, supported
}

// updatesHeading is a heading that displays number of updates available,
// whether updates are installing, and an Update All button.
type updatesHeading struct {
	vecty.Core
	Available uint
	Updating  bool

	// TODO: Find a place for this.
	UpdateSupported bool
}

func (u *updatesHeading) Render() *vecty.HTML {
	if u.Available == 0 && !u.Updating {
		return nil
	}
	return elem.Heading4(
		vecty.Property(atom.Style.String(), "text-align: left;"),
		vecty.If(u.Updating, vecty.Text("Updates Installing...")),
		vecty.If(!u.Updating, vecty.Text(fmt.Sprintf("%d Updates Available", u.Available))),
		elem.Span(
			vecty.Property(atom.Style.String(), "float: right;"),
			u.updateAllButton(),
		),
	)
}

func (u *updatesHeading) updateAllButton() *vecty.HTML {
	if !u.UpdateSupported {
		return elem.Span(
			vecty.Property(atom.Style.String(), "color: gray; cursor: default;"),
			vecty.Property(atom.Title.String(), "Updating repos is not currently supported for this source of repos."),
			vecty.Text("Update All"),
		)
	}
	switch {
	case u.Available > 0:
		return elem.Anchor(
			prop.Href("/api/update-all"), // TODO: Should it be a separate endpoint or what?
			event.Click(func(e *vecty.Event) {
				// TODO.
				fmt.Println("UpdateAll()")
				js.Global.Get("UpdateAll").Invoke() // TODO: Do this via action?
			}).PreventDefault(),
			vecty.Text("Update All"),
		)
	case u.Available == 0:
		return elem.Span(
			vecty.Property(atom.Style.String(), "color: gray; cursor: default;"),
			vecty.Text("Update All"),
		)
	default:
		panic("unreachable")
	}
}

// InstalledUpdates is a heading for installed updates.
func InstalledUpdates() *vecty.HTML { return heading(elem.Heading3, "Installed Updates") }

func checkingForUpdates() *vecty.HTML { return heading(elem.Heading2, "Checking for updates...") }

func noUpdates() *vecty.HTML { return heading(elem.Heading2, "No Updates Available") }

func heading(heading func(markup ...vecty.MarkupOrComponentOrHTML) *vecty.HTML, text string) *vecty.HTML {
	return heading(
		vecty.Property(atom.Style.String(), "text-align: center;"),
		vecty.Text(text),
	)
}
