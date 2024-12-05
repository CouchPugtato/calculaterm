package modules

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

/* Colors for expressions, cycles through:
* Blue -> Red -> Green -> Purple
 */

var possibleColors []tcell.Color
var checkboxColor = tcell.NewRGBColor(160, 160, 175)

func nextColor() tcell.Color {
	if len(possibleColors) == 0 {
		possibleColors = []tcell.Color{
			tcell.ColorBlue,
			tcell.ColorRed,
			tcell.ColorGreen,
			tcell.ColorPink,
		}
	}
	color := possibleColors[0]
	possibleColors = possibleColors[1:] // double check to make sure this does as intended
	return color
}

// returns a duller version of the inputted color, to show focus
// TODO: Make work later if still needed
func backgroundColor(color tcell.Color) tcell.Color {
	r, g, b := color.RGB()
	return tcell.NewRGBColor(r, g, b)
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
var queResponseUpdate = true
var lastFieldSize = 0
var queRemove = -1

// ExpressionBox is the main container for all 'Expression Field' elements, [fuctionality being] mainly for
var ExpressionBox = tview.NewFlex().SetDirection(tview.FlexColumnCSS).
	AddItem((func() *tview.Flex {
		newExpression(0)
		return Expressions[0].full
	})(), 0, 1, false)

// Updates BEFORE frame is drawn, returns true if drawing should not occur
func ExpressionsUpdate() bool {
	if queGraphUpdate {
		RedrawGraph()
		queGraphUpdate = false
	}
	if queRemove != -1 {
		removeExpression(queRemove)
		queRemove = -1
	}
	if queInputUpdate {
		updateExpressionBox()
		queInputUpdate = false
	}

	_, _, newFieldSize, _ := Expressions[0].responseField.GetRect()
	if queResponseUpdate || lastFieldSize != newFieldSize {
		maintainResponses()
		lastFieldSize = newFieldSize
		queResponseUpdate = false
	}
	return false
}

func updateExpressionBox() {
	var label strings.Builder
	for index, expr := range Expressions {
		// Changing the indexes of the expressions locally to match the real values
		expr.index = index
		if index == focusedExpressionIndex {
			expr.expressionField.SetBackgroundColor(expr.color)
		} else {
			// TODO: not working
			expr.expressionField.SetBackgroundColor(backgroundColor(expr.color))
		}
		expr.responseField.SetFieldBackgroundColor(tcell.ColorWhite)

		// Changes the expression names to be
		label.Reset()
		label.WriteString(" y")
		label.WriteString(strconv.Itoa(index + 1))
		label.WriteString(" =")
		expr.expressionField.SetLabel(label.String())
	}

	// Refreshes ExpressionBox with updated expressions
	ExpressionBox.Clear()
	for _, expr := range Expressions {
		ExpressionBox.AddItem(expr.full, 0, 1, false)
	}
}

// Correct any changes made to the response mesages
func maintainResponses() {
	var spacer strings.Builder
	for _, expr := range Expressions {
		expr.responseField.SetText(expr.responseText)
		expr.responseField.SetText(expr.responseText)

		_, _, width, _ := expr.responseField.GetRect()
		spacer.Reset()
		for i := 0; i < (width - utf8.RuneCountInString(expr.responseText)); i++ {
			spacer.WriteRune(' ')
		}
		expr.responseField.SetLabel(spacer.String())
	}
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
		SetLabelWidth(0).SetLabelColor(tcell.ColorBlack). // Use label to space response size
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorBlack).
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
		AddItem(responseField, 0, 1, false)

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
			responseText:    "Testing today",
			enabledCheckbox: enabledCheckbox,
		}},
			Expressions[index:]...)...)

	Expressions[index].enabledCheckbox.SetChangedFunc(func(checked bool) {
		Expressions[index].isEnabled = checked
		if !Expressions[index].hasError {
			RedrawGraph()
		}
	})
	Expressions[index].responseField.SetChangedFunc(func(text string) {
		queResponseUpdate = true
	})
	Expressions[index].expressionField.SetChangedFunc(func(text string) {
		Expressions[index].formationString = text
		Expressions[index].function, Expressions[index].hasError = calculateExpression(text)
		queGraphUpdate = true
	}).SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			Expressions[index].expressionField.SetText(formatExpressionText(Expressions[index].formationString))
			newExpression(focusedExpressionIndex + 1)
		case tcell.KeyEscape:
			//
		case tcell.KeyTab:
			// possibly link to autocomplete function
		case tcell.KeyBacktab:
			if Expressions[index].expressionField.GetText() == "" && Expressions[index].index != 0 {
				queRemove = Expressions[index].index
			}
		}
	}).SetFocusFunc(func() {
		focusedExpressionIndex = Expressions[index].index
	})

	queInputUpdate = true
	queResponseUpdate = true
}

func removeExpression(index int) {
	for i := len(Expressions); i > index; i-- {
		renameVariable(Expressions[i].name, Expressions[i-1].name)
		// add something to replace all instances of expression at index with the text inside of that expression
	}

	Expressions = append(
		Expressions[:index],
		Expressions[index+1:]...)
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
	/*
		terms := strings.Split(text, "")
		for i, term := range terms {

		}
	*/
	return func(x float64) float64 { return 0 }, false
}

// split exponentials by ^
