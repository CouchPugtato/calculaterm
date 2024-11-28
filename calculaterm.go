package main

import (
	"github.com/CouchPugtato/calculaterm/modules"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	flex := tview.NewFlex().
		AddItem(modules.Expressions(), 0, 0, false).
		AddItem(tview.NewFlex().
			AddItem(modules.Graph(), 0, 0, false).
			AddItem(modules.Information(), 0, 0, false), 0, 0, false)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
