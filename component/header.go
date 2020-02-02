package component

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
	"github.com/gopherjs/vecty/style"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
	"golang.org/x/net/html/atom"
)

// Header is a component that displays the header with tabs on top.
type Header struct {
	vecty.Core
}

// Render renders the component.
func (*Header) Render() vecty.ComponentOrHTML {
	return elem.Header(
		vecty.Markup(style.Width("100%"), vecty.Style("text-align", "center")),
		elem.Span(
			vecty.Markup(vecty.Style("padding", string(style.Px(15))), vecty.Style("display", "inline-block")),
			vecty.Text("Updates"),
		),
	)
}

// updatesHeader combines checkingForUpdates, noUpdates and updatesHeading
// into one high level component.
type updatesHeader struct {
	Active          []*model.RepoPresentation
	CheckingUpdates bool
}

func (u updatesHeader) Render() []vecty.MarkupOrChild {
	var ns []vecty.MarkupOrChild
	switch {
	case u.CheckingUpdates:
		// Show "Checking for updates..." while still checking.
		ns = append(ns, heading(elem.Heading2, "Checking for updates..."))
	case !u.CheckingUpdates && len(u.Active) == 0:
		// Show "No Updates Available" if we're done checking and there are no remaining updates.
		ns = append(ns,
			elem.Heading2(
				vecty.Markup(vecty.Style("text-align", "center"), vecty.Style("margin-bottom", "2px")),
				vecty.Text("No Updates Available"),
			),
			elem.Heading4(
				vecty.Markup(vecty.Style("text-align", "center"), vecty.Style("margin-top", "2px"), vecty.Style("font-weight", "normal")),
				vecty.Text("All your Go packages are up to date"),
			),
		)
	}
	// Show number of updates available and Update All button.
	available, updating, supported := u.status()
	ns = append(ns, &updatesHeading{
		Available:       available,
		Updating:        updating,
		UpdateSupported: supported, // TODO: Fetch this value from backend once.
	})
	return ns
}

// status reports available, updating, supported updates in u.Active.
func (u updatesHeader) status() (available uint, updating bool, supported bool) {
	for _, rp := range u.Active {
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
	Available uint `vecty:"prop"`
	Updating  bool `vecty:"prop"`

	// TODO: Find a place for this.
	UpdateSupported bool `vecty:"prop"`
}

func (u *updatesHeading) Render() vecty.ComponentOrHTML {
	if u.Available == 0 && !u.Updating {
		return nil
	}
	var status string
	if u.Available > 0 {
		status = fmt.Sprintf("%d Updates Available", u.Available)
	}
	if u.Updating {
		if status != "" {
			status += ", "
		}
		status += "Installing Updates..."
	}
	return elem.Heading4(
		vecty.Markup(vecty.Style("text-align", "left")),
		vecty.Text(status),
		elem.Span(
			vecty.Markup(vecty.Style("float", "right")),
			u.updateAllButton(),
		),
	)
}

func (u *updatesHeading) updateAllButton() *vecty.HTML {
	if !u.UpdateSupported {
		return elem.Span(
			vecty.Markup(
				style.Color("gray"), vecty.Style("cursor", "default"),
				vecty.Property(atom.Title.String(), "Updating repos is not currently supported for this source of repos."),
			),
			vecty.Text("Update All"),
		)
	}
	switch {
	case u.Available > 0:
		return elem.Anchor(
			vecty.Markup(
				prop.Href("/api/update-all"), // TODO: Should it be a separate endpoint or what?
				event.Click(func(e *vecty.Event) {
					// TODO.
					fmt.Println("UpdateAll()")
					js.Global.Get("UpdateAll").Invoke() // TODO: Do this via action?
				}).PreventDefault(),
			),
			vecty.Text("Update All"),
		)
	case u.Available == 0:
		return elem.Span(
			vecty.Markup(style.Color("gray"), vecty.Style("cursor", "default")),
			vecty.Text("Update All"),
		)
	default:
		panic("unreachable")
	}
}

func heading(heading func(markup ...vecty.MarkupOrChild) *vecty.HTML, text string) *vecty.HTML {
	return heading(
		vecty.Markup(vecty.Style("text-align", "center")),
		vecty.Text(text),
	)
}
