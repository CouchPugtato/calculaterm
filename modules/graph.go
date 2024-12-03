package modules

import (
	"github.com/rivo/tview"
)

const GraphSize = 16

func Graph() *tview.TextView {
	return tview.NewTextView()
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
