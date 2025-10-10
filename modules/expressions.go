package modules

import (
	"errors"
	"fmt"
	"math"
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

// Expression struct containing all information pertaining to each entry
type expression struct {
	formationString string
	name            string
	isEnabled       bool
	err             error
	color           tcell.Color
	index           int
	function        func(x float64) (float64, error)
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
	defaultName := "y" + strconv.Itoa(index+1)
	var color = nextColor()

	exprField := tview.NewInputField().
		SetFieldTextColor(tcell.ColorWhite).
		SetFieldBackgroundColor(color).
		SetText(defaultName + " = ")

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
			name:            defaultName,
			isEnabled:       true,
			err:             nil,
			color:           color,
			index:           index,
			function:        func(x float64) (float64, error) { return 0, nil },
			full:            full,
			expressionField: exprField,
			responseField:   responseField,
			responseText:    "",
			enabledCheckbox: enabledCheckbox,
		}},
			Expressions[index:]...)...)

	Expressions[index].enabledCheckbox.SetChangedFunc(func(checked bool) {
		Expressions[index].isEnabled = checked
		if Expressions[index].err == nil {
			RedrawGraph()
		}
	})
	Expressions[index].responseField.SetChangedFunc(func(text string) {
		queResponseUpdate = true
	})
	Expressions[index].expressionField.SetChangedFunc(func(text string) {
		// Support editing name inline via definition syntax: name = expression
		raw := strings.TrimSpace(text)
		var defName string
		var rhs string
		if strings.Contains(raw, "=") {
			parts := strings.SplitN(raw, "=", 2)
			// Allow exactly one optional space immediately before '='.
			// Still disallow spaces within the identifier itself.
			eqPos := strings.Index(text, "=")
			leftSegment := text
			if eqPos != -1 {
				leftSegment = text[:eqPos]
			}

			// Count trailing spaces before '='
			trailingSpaces := 0
			for i := len(leftSegment) - 1; i >= 0 && i < len(leftSegment); i-- {
				if leftSegment[i] == ' ' {
					trailingSpaces++
				} else {
					break
				}
			}
			if trailingSpaces > 1 {
				// More than one space before '=' is not allowed
				// Position points at the second trailing space
				pos := eqPos - trailingSpaces + 1
				Expressions[index].err = &ExpressionError{"only one space allowed before '='", pos}
				Expressions[index].responseText = Expressions[index].err.Error()
				queResponseUpdate = true
				return
			}

			// Disallow spaces inside the identifier itself (excluding allowed trailing single space)
			coreLeft := strings.TrimRight(leftSegment, " ")
			if strings.Contains(coreLeft, " ") {
				// Find the first internal space position
				spacePos := strings.Index(coreLeft, " ")
				// Map to original text position
				Expressions[index].err = &ExpressionError{"identifier cannot contain spaces", spacePos}
				Expressions[index].responseText = Expressions[index].err.Error()
				queResponseUpdate = true
				return
			}

			defName = strings.TrimSpace(strings.ToLower(parts[0]))
			rhs = strings.TrimSpace(parts[1])
			if defName != "" && isValidIdentifier(defName) {
				// Update name and label dynamically
				old := Expressions[index].name
				Expressions[index].name = defName
				if old != defName {
					renameVariable(old, defName)
				}
			} else {
				// Invalid name, keep previous
				rhs = raw
			}
		} else {
			rhs = raw
		}

		Expressions[index].formationString = rhs
		Expressions[index].function, Expressions[index].err = CreateFunction(rhs)
		if Expressions[index].err == nil {
			Expressions[index].responseText = ""
			// Register this expression under its name for cross-reference
			nameKey := strings.ToLower(Expressions[index].name)
			userFunctions[nameKey] = Expressions[index].function
			queGraphUpdate = true
		} else {
			Expressions[index].responseText = Expressions[index].err.Error()
			// If invalid, remove from user functions to avoid stale references
			delete(userFunctions, strings.ToLower(Expressions[index].name))
		}
		queResponseUpdate = true
	}).SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			// Expressions[index].expressionField.SetText(formatExpressionText(Expressions[index].formationString)) // TODO: make format expressionText
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

func renameVariable(originalName string, newName string) {
	// Replace occurrences of the originalName with newName in all expressions
	// and re-create their functions to keep everything consistent.
	for i := range Expressions {
		// Skip the expression that's currently being renamed to avoid recursive change callbacks
		if i == focusedExpressionIndex {
			continue
		}

		// Update the visible input text for other expressions
		fieldText := Expressions[i].expressionField.GetText()
		fieldUpdated := strings.ReplaceAll(fieldText, originalName, newName)
		if fieldUpdated != fieldText {
			Expressions[i].expressionField.SetText(fieldUpdated)
		}

		// Update the stored RHS expression
		updated := strings.ReplaceAll(Expressions[i].formationString, originalName, newName)
		if updated != Expressions[i].formationString {
			Expressions[i].formationString = updated
			Expressions[i].function, Expressions[i].err = CreateFunction(updated)
			if Expressions[i].err == nil {
				Expressions[i].responseText = ""
				userFunctions[strings.ToLower(Expressions[i].name)] = Expressions[i].function
			} else {
				Expressions[i].responseText = Expressions[i].err.Error()
			}
		}
	}
	queResponseUpdate = true
	queGraphUpdate = true
}

type ExpressionError struct {
	Message  string
	Position int
}

func (e *ExpressionError) Error() string {
	return fmt.Sprintf("%s at position %d", e.Message, e.Position)
}

type TokenType int

const (
	NUMBER TokenType = iota
	OPERATOR
	FUNCTION
	LPAREN
	RPAREN
	VARIABLE
	CONSTANT
)

type Token struct {
	Type     TokenType
	Value    string
	Position int
}

// Operator precedence
var precedence = map[string]int{
	"+": 1,
	"-": 1,
	"*": 2,
	"/": 2,
	"^": 3,
}

// Mathematical constants
var mathConstants = map[string]float64{
	"e":   math.E,
	"pi":  math.Pi,
	"phi": math.Phi,
	"tau": 2 * math.Pi,
}

// Built in mathematical functions
var mathFuncs = map[string]func(float64) float64{
	"sin":  math.Sin,
	"cos":  math.Cos,
	"tan":  math.Tan,
	"sqrt": math.Sqrt,
	"ln":   math.Log,
	"exp":  math.Exp,
	"abs":  math.Abs,
	"asin": math.Asin,
	"acos": math.Acos,
	"atan": math.Atan,
	"d/dx": nil, // d/dx is handled as a special case, but still needs to be recognized as a "function"
}

func validateExpression(expr string) error {
	if len(strings.TrimSpace(expr)) == 0 {
		return errors.New("empty expression")
	}

	// Check for unclosed parentheses
	parenCount := 0
	for i, char := range expr {
		if char == '(' {
			parenCount++
		} else if char == ')' {
			parenCount--
			if parenCount < 0 {
				return &ExpressionError{"unmatched closing parenthesis", i}
			}
		}
	}
	if parenCount > 0 {
		return &ExpressionError{"unclosed parenthesis", len(expr) - 1}
	}

	// Check for consecutive operators
	prevIsOp := true // treat start of expression as if it follows an operator
	for i, char := range expr {
		isOp := strings.ContainsRune("+-*/^", char)
		if isOp && prevIsOp && char != '-' { // TODO: allow negative numbers
			return &ExpressionError{"consecutive operators", i}
		}
		prevIsOp = isOp
	}

	// Check for trailing operator
	lastChar := expr[len(expr)-1]
	if strings.ContainsRune("+-*/^", rune(lastChar)) {
		return &ExpressionError{"trailing operator", len(expr) - 1}
	}

	return nil
}

// Converts a string expression into tokens
func tokenize(expr string) ([]Token, error) {
	// Check if it's a definition
	if strings.Contains(expr, "=") {
		return nil, errors.New("definitions must be processed separately")
	}

	if err := validateExpression(expr); err != nil {
		return nil, err
	}

	var tokens []Token
	i := 0

	// Skip leading whitespace
	expr = strings.TrimSpace(expr)

	for i < len(expr) {
		char := string(expr[i])
		curPos := i

		// Skip whitespace
		if char == " " {
			i++
			continue
		}

		// Handle numbers
		if (char >= "0" && char <= "9") || char == "." {
			num := ""
			dotCount := 0
			startPos := i

			// Read the entire number
			for i < len(expr) && ((expr[i] >= '0' && expr[i] <= '9') || expr[i] == '.') {
				if expr[i] == '.' {
					dotCount++
					if dotCount > 1 {
						return nil, &ExpressionError{"multiple decimal points in number", i}
					}
				}
				num += string(expr[i])
				i++
			}

			if num == "." {
				return nil, &ExpressionError{"invalid number format", curPos}
			}

			tokens = append(tokens, Token{Type: NUMBER, Value: num, Position: startPos})

			// Skip any whitespace after the number
			for i < len(expr) && expr[i] == ' ' {
				i++
			}

			// Check if there's another number after whitespace
			if i < len(expr) {
				nextChar := expr[i]
				if (nextChar >= '0' && nextChar <= '9') || nextChar == '.' {
					return nil, &ExpressionError{"missing operator between numbers", i}
				}
			}
			continue
		}

		// Handle operators
		if char == "+" || char == "-" || char == "*" || char == "/" || char == "^" {
			tokens = append(tokens, Token{Type: OPERATOR, Value: char, Position: curPos})
			i++
			continue
		}

		// Handle parentheses
		if char == "(" {
			tokens = append(tokens, Token{Type: LPAREN, Value: char, Position: curPos})
			i++
			continue
		}
		if char == ")" {
			tokens = append(tokens, Token{Type: RPAREN, Value: char, Position: curPos})
			i++
			continue
		}

		// Handle variables
		if char == "x" {
			tokens = append(tokens, Token{Type: VARIABLE, Value: char, Position: curPos})
			i++
			continue
		}

		// Handle constants and functions
		if char >= "a" && char <= "z" {
			name := ""
			startPos := i
			// Special case for d/dx
			if i+3 < len(expr) && expr[i:i+4] == "d/dx" {
				name = "d/dx"
				i += 4
			} else {
				// Regular identifier handling
				for i < len(expr) && ((expr[i] >= 'a' && expr[i] <= 'z') || (expr[i] >= '0' && expr[i] <= '9')) {
					name += string(expr[i])
					i++
				}
			}

			// Check if it's a constant (built in or user defined)
			if _, exists := mathConstants[name]; exists {
				tokens = append(tokens, Token{Type: CONSTANT, Value: name, Position: startPos})
				continue
			}
			if _, exists := userConstants[name]; exists {
				tokens = append(tokens, Token{Type: CONSTANT, Value: name, Position: startPos})
				continue
			}

			// Check if it's a function (built in or user defined)
			if _, exists := mathFuncs[name]; exists {
				tokens = append(tokens, Token{Type: FUNCTION, Value: name, Position: startPos})
				continue
			}
			if _, exists := userFunctions[name]; exists {
				tokens = append(tokens, Token{Type: FUNCTION, Value: name, Position: startPos})
				continue
			}

			return nil, &ExpressionError{fmt.Sprintf("unknown identifier: %s", name), startPos}
		}

		return nil, &ExpressionError{"invalid character", i}
	}

	return tokens, nil
}

// handles user defined constants and functions
func processDefinition(expr string) error {
	// Special case for derivative expressions
	if strings.HasPrefix(strings.TrimSpace(expr), "f = d/dx") {
		// Extract the expression to be differentiated
		parts := strings.SplitN(expr, "d/dx", 2)
		if len(parts) != 2 {
			return errors.New("invalid derivative expression")
		}

		// Get the expression inside parentheses
		exprToDerive := strings.TrimSpace(parts[1])
		if !strings.HasPrefix(exprToDerive, "(") || !strings.HasSuffix(exprToDerive, ")") {
			return errors.New("derivative expression must be enclosed in parentheses")
		}

		// Remove the parentheses
		exprToDerive = exprToDerive[1 : len(exprToDerive)-1]

		// Calculate the derivative
		f, err := Derivative(exprToDerive)
		if err != nil {
			return fmt.Errorf("error calculating derivative: %v", err)
		}

		userFunctions["f"] = f
		return nil
	}

	// Original definition processing
	parts := strings.Split(expr, "=")
	if len(parts) != 2 {
		return errors.New("invalid definition format")
	}

	name := strings.TrimSpace(parts[0])
	// Disallow spaces in identifier and provide a clear error with position
	if strings.Contains(name, " ") {
		spacePos := strings.Index(expr, " ")
		if spacePos == -1 {
			spacePos = 0
		}
		return &ExpressionError{"identifier cannot contain spaces", spacePos}
	}
	value := strings.TrimSpace(parts[1])

	// Validate the name
	if len(name) == 0 {
		return errors.New("empty name in definition")
	}
	if !isValidIdentifier(name) {
		return fmt.Errorf("invalid identifier name: %s", name)
	}
	if name == "x" {
		return errors.New("cannot redefine variable 'x'")
	}
	if _, exists := mathFuncs[name]; exists {
		return fmt.Errorf("cannot redefine built-in function: %s", name)
	}
	if _, exists := mathConstants[name]; exists {
		return fmt.Errorf("cannot redefine built-in constant: %s", name)
	}
	if _, exists := userFunctions[name]; exists {
		return fmt.Errorf("function %s is already defined", name)
	}
	if _, exists := userConstants[name]; exists {
		return fmt.Errorf("constant %s is already defined", name)
	}

	// Check if it's a function definition
	isFunction := false
	tokens, err := tokenize(value)
	if err != nil {
		return fmt.Errorf("invalid expression in definition: %v", err)
	}
	for _, token := range tokens {
		if token.Type == VARIABLE {
			isFunction = true
			break
		}
	}

	if isFunction {
		f, err := CreateFunction(value)
		if err != nil {
			return fmt.Errorf("invalid function definition: %v", err)
		}
		userFunctions[name] = f
	} else {
		f, err := CreateFunction(value)
		if err != nil {
			return fmt.Errorf("invalid constant definition: %v", err)
		}
		userConstants[name], _ = f(0)
	}
	return nil
}

func isValidIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}
	// First character must be a letter
	if name[0] < 'a' || name[0] > 'z' {
		return false
	}
	// Rest can be letters or numbers
	for i := 1; i < len(name); i++ {
		if !((name[i] >= 'a' && name[i] <= 'z') || (name[i] >= '0' && name[i] <= '9')) {
			return false
		}
	}
	return true
}

// Convert infix notation to postfix
func toPostfix(tokens []Token) ([]Token, error) {
	var output []Token
	var stack []Token

	for _, token := range tokens {
		switch token.Type {
		case NUMBER, VARIABLE, CONSTANT:
			output = append(output, token)

		case FUNCTION:
			stack = append(stack, token)

		case OPERATOR:
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Type == OPERATOR && precedence[top.Value] >= precedence[token.Value] {
					output = append(output, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, token)

		case LPAREN:
			stack = append(stack, token)

		case RPAREN:
			foundMatching := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.Type == LPAREN {
					foundMatching = true
					if len(stack) > 0 && stack[len(stack)-1].Type == FUNCTION {
						output = append(output, stack[len(stack)-1])
						stack = stack[:len(stack)-1]
					}
					break
				}
				output = append(output, top)
			}
			if !foundMatching {
				return nil, &ExpressionError{"unmatched closing parenthesis", token.Position}
			}
		}
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.Type == LPAREN {
			return nil, &ExpressionError{"unclosed parenthesis", top.Position}
		}
		output = append(output, top)
	}

	return output, nil
}

var userFunctions = make(map[string]func(float64) (float64, error))
var userConstants = make(map[string]float64)

func CreateFunction(expr string) (func(float64) (float64, error), error) {
	// Check if it's a definition
	if strings.Contains(expr, "=") {
		if err := processDefinition(expr); err != nil {
			return nil, err
		}
		return nil, nil
	}

	tokens, err := tokenize(expr)
	if err != nil {
		return nil, err
	}

	// Add token sequence validation
	if err := validateTokenSequence(tokens); err != nil {
		return nil, err
	}

	postfix, err := toPostfix(tokens)
	if err != nil {
		return nil, err
	}

	return func(x float64) (float64, error) {
		var stack []float64

		for _, token := range postfix {
			switch token.Type {
			case NUMBER:
				val := 0.0
				fmt.Sscanf(token.Value, "%f", &val)
				stack = append(stack, val)

			case VARIABLE:
				stack = append(stack, x)

			case CONSTANT:
				stack = append(stack, mathConstants[token.Value])

			case OPERATOR:
				if len(stack) < 2 {
					return 0, errors.New("invalid expression: not enough operands")
				}
				b := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				a := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				var result float64
				switch token.Value {
				case "+":
					result = a + b
				case "-":
					result = a - b
				case "*":
					result = a * b
				case "/":
					if b == 0 {
						return 0, errors.New("division by zero")
					}
					result = a / b
				case "^":
					result = math.Pow(a, b)
					if math.IsNaN(result) || math.IsInf(result, 0) {
						return 0, errors.New("invalid power operation result")
					}
				}
				stack = append(stack, result)

			case FUNCTION:
				if len(stack) < 1 {
					return 0, errors.New("invalid expression: not enough arguments for function")
				}
				a := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				var result float64
				if token.Value == "d/dx" {
					// Get the original tokens that represent the inner expression
					var innerTokens []Token
					parenCount := 0
					var startIndex int

					// Find this specific d/dx function's tokens
					for i := 0; i < len(tokens); i++ {
						if tokens[i].Position == token.Position {
							startIndex = i + 1
							break
						}
					}

					// Collect tokens for this specific d/dx instance
					for i := startIndex; i < len(tokens); i++ {
						if tokens[i].Type == LPAREN {
							parenCount++
							if parenCount == 1 {
								continue // Skip the first opening parenthesis
							}
						} else if tokens[i].Type == RPAREN {
							parenCount--
							if parenCount == 0 {
								break // Found the matching closing parenthesis
							}
						}
						if parenCount > 0 {
							innerTokens = append(innerTokens, tokens[i])
						}
					}

					// Create a function from these tokens directly
					innerExpr := tokensToString(innerTokens)
					f, err := CreateFunction(innerExpr)
					if err != nil {
						return 0, err
					}

					result, err = NumericalDerivative(f, x)
					if err != nil {
						return 0, err
					}
				} else if f, exists := userFunctions[token.Value]; exists {
					result, _ = f(a)
				} else {
					// Handle built in functions
					switch token.Value {
					case "ln":
						if a <= 0 {
							return 0, fmt.Errorf("domain error: ln(%f) - logarithm of non-positive number", a)
						}
						result = math.Log(a)
					case "sqrt":
						if a < 0 {
							return 0, fmt.Errorf("domain error: sqrt(%f) - square root of negative number", a)
						}
						result = math.Sqrt(a)
					case "asin", "acos":
						if a < -1 || a > 1 {
							return 0, fmt.Errorf("domain error: %s(%f) - argument must be between -1 and 1", token.Value, a)
						}
						if token.Value == "asin" {
							result = math.Asin(a)
						} else {
							result = math.Acos(a)
						}
					default:
						result = mathFuncs[token.Value](a)
					}
				}

				// Check for general domain errors
				if math.IsNaN(result) || math.IsInf(result, 0) {
					return 0, fmt.Errorf("domain error: %s(%f) produced an invalid result", token.Value, a)
				}
				stack = append(stack, result)
			}
		}

		if len(stack) != 1 {
			return 0, errors.New("invalid expression: incorrect number of values in final stack")
		}

		return stack[0], nil
	}, nil
}

func validateTokenSequence(tokens []Token) error {
	if len(tokens) == 0 {
		return errors.New("empty token sequence")
	}

	for i := 0; i < len(tokens)-1; i++ {
		curr := tokens[i]
		next := tokens[i+1]

		// Check for invalid sequences
		switch curr.Type {
		case OPERATOR:
			if next.Type == OPERATOR {
				return &ExpressionError{"consecutive operators", next.Position}
			}
		case NUMBER:
			if next.Type == NUMBER {
				return &ExpressionError{"missing operator between numbers", next.Position}
			}
			// Allow implicit multiplication between number and variable/function/constant
			if next.Type != OPERATOR && next.Type != RPAREN {
				continue
			}
		case FUNCTION:
			if next.Type != LPAREN {
				return &ExpressionError{"missing opening parenthesis after function", next.Position}
			}
		case LPAREN:
			if next.Type == RPAREN {
				return &ExpressionError{"empty parentheses", next.Position}
			}
			if next.Type == OPERATOR {
				return &ExpressionError{"operator after opening parenthesis", next.Position}
			}
		case RPAREN:
			if next.Type == LPAREN || next.Type == FUNCTION || next.Type == VARIABLE || next.Type == CONSTANT || next.Type == NUMBER {
				return &ExpressionError{"missing operator after closing parenthesis", next.Position}
			}
		case VARIABLE:
			// Allow implicit multiplication between variable and number/function/constant
			if next.Type != OPERATOR && next.Type != RPAREN {
				continue
			}
		case CONSTANT:
			// Allow implicit multiplication between constant and number/variable/function/constant
			if next.Type != OPERATOR && next.Type != RPAREN {
				continue
			}
		}
	}

	// Check first and last tokens
	if len(tokens) > 0 {
		first := tokens[0]
		last := tokens[len(tokens)-1]

		if first.Type == OPERATOR && first.Value != "-" {
			return &ExpressionError{"expression cannot start with operator", first.Position}
		}
		if last.Type == OPERATOR {
			return &ExpressionError{"expression cannot end with operator", last.Position}
		}
	}

	return nil
}

// Evaluates the derivative at a point using a fourth-order central difference formula
func NumericalDerivative(f func(float64) (float64, error), x float64) (float64, error) {
	// Calculate optimal step size based on x
	h := math.Pow(2.2e-16, 1.0/5.0) * math.Max(1.0, math.Abs(x))

	a, err1 := f(x + 2*h)
	b, err2 := f(x + h)
	c, err3 := f(x - h)
	d, err4 := f(x - 2*h)

	if _, err0 := f(x); err0 != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return 0, errors.New("Derivative does not exist at this point")
	}
	// Fourth-order central difference formula
	return (-a + 8*b - 8*c + d) / (12 * h), nil
}

func Derivative(expr string) (func(float64) (float64, error), error) {
	f, err := CreateFunction(expr)
	if err != nil {
		return nil, err
	}

	// Return a new function that uses numerical differentiation on inner expression function
	return func(x float64) (float64, error) {
		return NumericalDerivative(f, x)
	}, nil
}

func tokensToString(tokens []Token) string {
	var result strings.Builder
	for i, token := range tokens {
		if i > 0 && token.Type != RPAREN && tokens[i-1].Type != LPAREN {
			result.WriteString(" ")
		}
		result.WriteString(token.Value)
	}
	return result.String()
}
