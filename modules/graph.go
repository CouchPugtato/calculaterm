package modules

import (
	"github.com/rivo/tview"
)

const GraphSize = 16

var Graph = tview.NewTextView()

// Temperary function to print into graph
func GraphPrint(a string) {
	Graph.SetText(Graph.GetText(true) + "\n" + a)
}

/*
func graphMouseHandling(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {

}
*/
func GraphTraversial() *tview.Flex {
	/*
		buttons needed:
			- Home button
			- zoom out
			- zoom in
			- up, right, left, down
	*/

	return tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(true), 0, 1, false)
}

func GraphUpdate() bool {
	// Updates BEFORE frame is drawn, returns true if drawing should not occur
	return false
}

func RedrawGraph() {

}
