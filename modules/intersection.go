package modules

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var IntersectionBox *tview.Flex
var intersectInput1 *tview.InputField
var intersectInput2 *tview.InputField
var intersectGuess *tview.InputField
var intersectResult *tview.TextView

func init() {
	intersectInput1 = tview.NewInputField().
		SetLabel("f(x) = ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

	intersectInput2 = tview.NewInputField().
		SetLabel("g(x) = ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

	intersectGuess = tview.NewInputField().
		SetLabel("Guess x = ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

	intersectResult = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	calcBtn := tview.NewButton("Find Intersection").
		SetSelectedFunc(calculateIntersection)
	calcBtn.SetBackgroundColor(tcell.ColorDarkGray)
	calcBtn.SetLabelColor(tcell.ColorWhite)

	IntersectionBox = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(intersectInput1, 1, 1, false).
		AddItem(intersectInput2, 1, 1, false).
		AddItem(intersectGuess, 1, 1, false).
		AddItem(calcBtn, 1, 1, false).
		AddItem(intersectResult, 2, 1, false)

	IntersectionBox.SetBorder(true).SetTitle("Intersection Finder")
}

func calculateIntersection() {
	expr1 := intersectInput1.GetText()
	expr2 := intersectInput2.GetText()
	guessStr := intersectGuess.GetText()

	if expr1 == "" || expr2 == "" || guessStr == "" {
		intersectResult.SetText("[red]Please fill all fields")
		return
	}

	f1, err1 := CreateFunction(expr1)
	if err1 != nil {
		intersectResult.SetText("[red]f(x) Error: " + err1.Error())
		return
	}
	f2, err2 := CreateFunction(expr2)
	if err2 != nil {
		intersectResult.SetText("[red]g(x) Error: " + err2.Error())
		return
	}

	guess, err := strconv.ParseFloat(guessStr, 64)
	if err != nil {
		intersectResult.SetText("[red]Invalid guess")
		return
	}

	// Define the difference function
	diffFunc := func(x float64) (float64, error) {
		v1, e1 := f1(x)
		if e1 != nil {
			return 0, e1
		}
		v2, e2 := f2(x)
		if e2 != nil {
			return 0, e2
		}
		return v1 - v2, nil
	}

	root, err := findRoot(diffFunc, guess)
	if err != nil {
		intersectResult.SetText("[red]Error: " + err.Error())
	} else {
		intersectResult.SetText(fmt.Sprintf("[green]Intersection at x = %.4f", root))
	}
}

func findRoot(f func(float64) (float64, error), guess float64) (float64, error) {
	// Secant method
	x0 := guess
	x1 := guess + 0.1

	y0, err := f(x0)
	if err != nil {
		return 0, err
	}
	if math.Abs(y0) < 1e-9 {
		return x0, nil
	}

	maxIter := 100
	for i := 0; i < maxIter; i++ {
		y1, err := f(x1)
		if err != nil {
			return 0, err
		}

		if math.Abs(y1) < 1e-9 {
			return x1, nil
		}

		if math.Abs(y1-y0) < 1e-15 { // Avoid division by zero
			return 0, fmt.Errorf("no convergence (flat gradient)")
		}

		x2 := x1 - y1*(x1-x0)/(y1-y0)

		x0 = x1
		y0 = y1
		x1 = x2
	}

	return 0, fmt.Errorf("failed to converge after %d iterations", maxIter)
}
