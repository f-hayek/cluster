package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	MainColor          = tcell.ColorOrange
	TextColor          = tcell.ColorWhite
	TopbarTextColor    = tcell.ColorBlack
	StatusbarBgColor   = tcell.ColorBlack
	StatusbarTextColor = tcell.ColorWhite
)

type UI struct {
	app        *tview.Application
	pages      *tview.Pages
	primitives map[string]tview.Primitive
	menu       *tview.List
	log        *Log
	rpcPath    string
}

func NewTopBar() tview.Primitive {
	topbar := tview.NewTextView()
	topbar.SetBorder(false)
	topbar.SetText(" Cluster v0.1 - Press h for help")
	topbar.SetTextColor(TopbarTextColor)
	topbar.SetBackgroundColor(MainColor)
	return topbar
}
func NewStatusBar() tview.Primitive {
	bar := tview.NewTextView()
	bar.SetBorder(false)
	bar.SetText("  [j/k] Down/Up    [G/g] Bottom/top    [Enter] Details    [ESC] Back")
	bar.SetTextColor(StatusbarTextColor)
	bar.SetBackgroundColor(StatusbarBgColor)
	return bar
}

func NewMenu(ui *UI) *tview.List {
	menu := tview.NewList().ShowSecondaryText(false).
		AddItem("Node info", "Display general information about this node", 'i', func() {
			ui.log.Info("[green]Dash")
			ui.pages.SwitchToPage("dash")
		}).
		AddItem("Pay", "Pay an invoice", 'p', func() {
			ui.pages.SwitchToPage("pay")
			ui.SetFocus("pay")
		}).
		AddItem("Receive", "Receive funds", 'r', func() {
			if ui.pages.HasPage("receive") {
				ui.pages.SwitchToPage("receive")
			} else {
				ui.AddPage("receive", receivePage(ui), true, true)
				ui.pages.SwitchToPage("receive")
			}
			ui.SetFocus("receive")
		}).
		AddItem("Channels", "Display a list of all channels", 'c', func() {
			if ui.pages.HasPage("channels") {
				ui.pages.SwitchToPage("channels")
			} else {
				ui.AddPage("channels", channelsPage(ui), true, true)
				ui.pages.SwitchToPage("channels")
			}
			ui.SetFocus("channels")
		}).
		AddItem("Modal", "", 'm', func() {
			ui.pages.SwitchToPage("modal")
		}).
		AddItem("Quit", "Press to exit", 'q', func() {
			ui.app.Stop()
		})
	menu.SetBorder(true)
	menu.SetBorderColor(MainColor)
	menu.SetTitle(" Menu ")

	return menu

}

func (ui *UI) NewDashPage() tview.Primitive {
	dash := tview.NewTextView()
	dash.SetBorder(true)
	dash.SetBorderColor(MainColor)
	dash.SetTitle(" Dashboard ")
	dash.SetText("Hello")
	return dash
}

func (ui *UI) NewHelpPage() tview.Primitive {
	tv := tview.NewTextView()
	tv.SetBorderColor(MainColor)
	tv.SetBorder(true)
	tv.SetTitle(" HELP ")
	tv.SetText("Some help text")
	tv.SetDoneFunc(func(key tcell.Key) {
		ui.pages.HidePage("help")
	})
	return ui.Modal(tv, 80, 40)
}

func (ui *UI) AddPage(name string, p tview.Primitive, resize, visible bool) tview.Primitive {
	ui.pages.AddPage(name, p, resize, visible)
	ui.primitives[name] = p
	return p
}
func (ui *UI) SetFocus(name string) tview.Primitive {
	p := ui.primitives[name]
	ui.app.SetFocus(p)
	return p
}
func (ui *UI) FocusMenu() {
	ui.app.SetFocus(ui.menu)
}
func (ui *UI) SetupPages() *tview.Pages {
	ui.AddPage("dash", ui.NewDashPage(), true, true)
	ui.AddPage("pay", payPage(ui), true, false)
	ui.AddPage("help", ui.NewHelpPage(), true, false)
	return ui.pages
}
func (ui *UI) NewLayout() tview.Primitive {
	page := tview.NewGrid()
	page.SetColumns(30, 0)
	page.SetRows(1, 0, 1, 7)

	topBar := NewTopBar()

	ui.menu = NewMenu(ui)

	ui.log.view.SetChangedFunc(func() {
		ui.app.ForceDraw()
	})
	status := NewStatusBar()

	page.AddItem(topBar, 0, 0, 1, 2, 0, 0, true)
	page.AddItem(ui.menu, 1, 0, 2, 1, 0, 0, false)
	page.AddItem(ui.pages, 1, 1, 1, 1, 0, 0, false)
	page.AddItem(status, 2, 1, 1, 1, 0, 0, false)
	page.AddItem(ui.log.view, 3, 0, 1, 2, 0, 0, false)
	return page

}
func (ui *UI) Run() {
	layout := ui.NewLayout()
	ui.SetupPages()
	if err := ui.app.SetRoot(layout, true).SetFocus(ui.menu).Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) Modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}
