package main

import (
	"github.com/CouchPugtato/calculaterm/modules"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		return modules.ExpressionsUpdate() &&
			modules.GraphUpdate() &&
			modules.InformationUpdate()
	})

	graph := tview.NewFlex().SetDirection(tview.FlexColumnCSS).
		AddItem(tview.NewTextArea().SetBorder(true), 0, modules.GraphSize, false).
		AddItem(tview.NewTextArea().SetBorder(true), 0, modules.GraphSize/3, false).
		AddItem(tview.NewTextArea().SetBorder(true), 0, modules.GraphSize/2, false)

	full := tview.NewFlex().
		AddItem(modules.ExpressionBox, 0, 1, false).
		AddItem(graph, 0, 1, false)

	if err := app.SetRoot(full, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
