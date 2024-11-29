package main

import (
	"github.com/CouchPugtato/calculaterm/modules"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	graph := tview.NewFlex().SetDirection(tview.FlexColumnCSS).
		AddItem(tview.NewTextArea().SetBorder(true), 0, modules.GraphSize, false).
		AddItem(tview.NewTextArea().SetBorder(true), 0, modules.GraphSize/3, false).
		AddItem(tview.NewTextArea().SetBorder(true), 0, modules.GraphSize/2, false)

	full := tview.NewFlex().
		AddItem(modules.Expressions, 0, 1, false).
		AddItem(graph, 0, 1, false)

	if err := app.SetRoot(full, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
