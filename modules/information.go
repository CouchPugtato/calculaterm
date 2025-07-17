package modules

import (
	"github.com/rivo/tview"
)

var Information = tview.NewTextArea()

func InformationUpdate() bool {
	// Updates BEFORE frame is drawn, returns true if drawing should not occur
	return false
}

// Temporary function to print into graph
func InfoPrint(a string) {
	Information.SetText(Information.GetText()+"\n"+a, false)
}
