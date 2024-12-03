package modules

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

/* Colors for expressions, cycles through:
* Blue -> Red -> Green -> Purple -> Yellow -> Pink
 */

var possibleColors []tcell.Color
var checkboxColor = tcell.NewRGBColor(160, 160, 175)

func nextColor() tcell.Color {
	if len(possibleColors) == 0 {
		possibleColors = []tcell.Color{
			tcell.ColorBlue,
			tcell.ColorRed,
			tcell.ColorGreen,
			tcell.ColorYellow,
			tcell.ColorPink,
		}
	}
	color := possibleColors[0]
	possibleColors = possibleColors[1:] // double check to make sure this does as intended
	return color
}

func backgroundColor(color tcell.Color) tcell.Color {
	r, g, b := color.RGB()
	return tcell.NewRGBColor(r, g, b).TrueColor()
}

// expression strct declaration, used for containing...
type expression struct {
	formationString string
	name            string
	isEnabled       bool
	hasError        bool
	color           tcell.Color
	index           int
	function        func(x float64) float64
	full            *tview.Flex
	expressionField *tview.InputField
	responseField   *tview.InputField
	responseText    string
	enabledCheckbox *tview.Checkbox
}

var Expressions = []expression{}
var focusedExpressionIndex = 0
var queInputUpdate = false
var queGraphUpdate = false
var queRemove = -1

// ExpressionBox is the main container for all 'Expression Field' elements, [fuctionality being] mainly for
var ExpressionBox = tview.NewFlex().SetDirection(tview.FlexColumnCSS).
	AddItem((func() *tview.Flex {
		newExpression(0)
		return Expressions[0].full
	})(), 0, 1, false)

// Updates BEFORE frame is drawn, returns true if drawing should not occur
func ExpressionsUpdate() bool {
	if queInputUpdate {
		updateExpressionBox()
		queInputUpdate = false
	}
	if queGraphUpdate {
		RedrawGraph()
		queGraphUpdate = false
	}
	if queRemove != -1 {
		removeExpression(queRemove)
		queRemove = -1
	}
	return false
}

func updateExpressionBox() {
	for index, expr := range Expressions {
		expr.index = index
		if index == focusedExpressionIndex {
			expr.expressionField.SetBackgroundColor(expr.color)
		} else {
			expr.expressionField.SetBackgroundColor(backgroundColor(expr.color))
		}
	}

	ExpressionBox.Clear()
	for _, expr := range Expressions {
		ExpressionBox.AddItem(expr.full, 0, 1, false)
	}

	GraphPrint("expressionboxUpdated")
}

// Creates a new Expression, inserted at the index
func newExpression(index int) {
	/*- Creates a new 'Expression Field'
	[structure]:
		- Flex "Container" (CSS Collumn)
			- Flex "Container" (CSS Row)
				- Enable/Disable Function Checkbox, whether or not the function should be displayed on the graph/information
				- Function input field, responsible for the collection of the function generating text
			- Flex "Container" (CSS Row)
				- Empty Box used for spacing
				- Read-only input field, used for reporting information back to the user
	*/
	var label strings.Builder
	label.WriteString(" y")
	label.WriteString(strconv.Itoa(index + 1))
	label.WriteString(" =")

	var color = nextColor()

	exprField := tview.NewInputField().
		SetLabel(label.String()).
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(color)

	responseField := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorWhite).
		SetFieldTextColor(tcell.ColorBlack.TrueColor()).
		SetText("testing message")

	enabledCheckbox := tview.NewCheckbox().
		SetFieldBackgroundColor(checkboxColor).
		SetFieldTextColor(tcell.ColorBlack).
		SetChecked(true)

	full := tview.NewFlex().SetDirection(tview.FlexColumnCSS).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRowCSS).
			AddItem(enabledCheckbox, 1, 1, false).
			AddItem(exprField, 0, 20, false),
			1, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRowCSS).
			AddItem(tview.NewBox(), 0, 4, false).
			AddItem(responseField, 0, 1, false), 0, 1, false)

	Expressions = append(
		Expressions[:index],
		append([]expression{{
			formationString: "",
			name:            label.String(),
			isEnabled:       true,
			hasError:        false,
			color:           color,
			index:           index,
			function:        func(x float64) float64 { return 0 },
			full:            full,
			expressionField: exprField,
			responseField:   responseField,
			responseText:    "",
			enabledCheckbox: enabledCheckbox,
		}},
			Expressions[index:]...)...)

	Expressions[index].enabledCheckbox.SetChangedFunc(func(checked bool) {
		Expressions[index].isEnabled = checked
		if !Expressions[index].hasError {
			RedrawGraph()
		}
		GraphPrint("Checked")
	})
	Expressions[index].responseField.SetChangedFunc(func(text string) {
		// Error when trying to change, causes application to crash
		Expressions[index].responseField.SetText(Expressions[index].responseText)
	})
	Expressions[index].expressionField.SetChangedFunc(func(text string) {
		Expressions[index].formationString = formatExpressionText(text)
		Expressions[index].function, Expressions[index].hasError = calculateExpression(text)
		queGraphUpdate = true
	}).SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			newExpression(focusedExpressionIndex + 1)
		case tcell.KeyEscape:
			//
		case tcell.KeyTab:
			// possibly link to autocomplete function
		case tcell.KeyBackspace:
			if Expressions[index].formationString == "" && Expressions[index].index != 0 {
				queRemove = Expressions[index].index
			}
		}
	}).SetFocusFunc(func() {
		focusedExpressionIndex = Expressions[index].index
	})

	queInputUpdate = true
	GraphPrint("Ran for index: " + strconv.Itoa(index) + "; ExprLen: " + strconv.Itoa(len(Expressions)))
}

// Removes an expression f
func removeExpression(index int) {
	Expressions = append(
		Expressions[:index],
		Expressions[index:]...)
	updateExpressionBox()
}

// For replacement of characters and for formatting purposes
func formatExpressionText(text string) string {
	// TODO: replace later
	return text
}

func renameVariable(originalName string, newName string) {
	// go through all function generating strings and replace variables to thier new names
}

// Expression Evaluation --------------------------------------------

// Returns a generating function
func calculateExpression(text string) (func(x float64) float64, bool) {
	return func(x float64) float64 { return 0 }, false
}
