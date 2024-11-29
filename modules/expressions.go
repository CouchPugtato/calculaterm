package modules

import (
	"github.com/rivo/tview"
)

const boxHeight = 0

var Expressions *tview.Flex = tview.NewFlex().
	SetDirection(tview.FlexColumnCSS).
	AddItem(tview.NewInputField().SetBorder(true), 0, 1, true)
