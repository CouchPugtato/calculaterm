package modules

import (
	"github.com/rivo/tview"
)

const GraphSize = 16 // graph proportionality in the flex that it is inside of, all other components scale off of it
var lastGraphSize = [2]int{0, 0}

const graphLiner = '╋'
const graphSpacer = '━'

var Graph = tview.NewTextView().
	SetText("╋╳╳╳╳╳╳╳╳╳╳╳\n╋╳╳╳╳╳╳╳╳╳╳\n╋╳╳╳╳╳╳╳╳╳╳╳╳\n╋╳╳╳╳╳╳╳╳╳╳╳╳╳\n╋╳╳╳╳╳╳╳╳╳╳╳\n╋╳╳╳╳╳╳╳╳╳╳╳\n╋╳╳╳╳╳╳╳╳╳╳╳╳\n╋╳╳╳╳╳╳╳╳╳╳╳╳\n╋━╋━╋━╋━╋━╋")

// Temporary function to print into graph
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

	_, _, newWidth, newHeight := Graph.GetRect()
	if newWidth != lastGraphSize[0] || newHeight != lastGraphSize[1] {
		lastGraphSize[0] = newWidth
		lastGraphSize[1] = newHeight
		RedrawGraph()
	}

	return false
}

func RedrawGraph() {

}
