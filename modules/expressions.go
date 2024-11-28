package modules

import (
	"github.com/rivo/tview"
)

func Expressions() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	return flex
}
