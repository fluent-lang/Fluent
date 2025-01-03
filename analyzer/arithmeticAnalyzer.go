package analyzer

import (
	"fluent/code"
	"fluent/code/mod"
	"fluent/code/wrapper"
	"fluent/logger"
	"fluent/stack"
	"fluent/token"
)

var arithmeticOperators = map[token.Type]struct{}{
	token.Plus:     {},
	token.Minus:    {},
	token.Asterisk: {},
	token.Slash:    {},
}

// isOperator checks if the given token is an arithmetic operator
func isOperator(token token.Token) bool {
	_, found := arithmeticOperators[token.GetType()]
	return found
}

// checkOperator checks if the given token is an arithmetic operator
func checkOperator(token token.Token, statement []token.Token) {
	if !isOperator(token) || len(statement) == 1 {
		logger.TokenError(
			token,
			"Invalid arithmetic expression",
			"An arithmetic expression must start with a number or a variable",
			"Check the expression",
		)
	}
}

// checkStatement checks if the given statement is a valid arithmetic expression
func checkStatement(
	extractingIdentifier *bool,
	lastStatement *[]token.Token,
	currentToken token.Token,
	currentIndex int,
	variables *stack.Stack,
	functions *map[string]map[string]*code.Function,
	mods *map[string]map[string]*mod.FluentMod,
) {
	*extractingIdentifier = false

	// Check the last statement
	if len(*lastStatement) == 0 {
		logger.TokenError(
			currentToken,
			"Unexpected token",
			"Expected an identifier or a number",
			"Check the arithmetic expression",
		)
	}

	// Analyze the last statement
	dummyObj := wrapper.NewFluentObject(
		dummyNothingType,
		nil,
	)

	dummyBool := false
	isArithmetic := false

	AnalyzeIdentifier(
		*lastStatement,
		variables,
		functions,
		mods,
		&currentIndex,
		&dummyObj,
		&isArithmetic,
		&dummyBool,
	)

	if !isArithmetic {
		logger.TokenError(
			currentToken,
			"Invalid arithmetic expression",
			"Expected an arithmetic expression",
			"Check the arithmetic expression",
		)
	}

	// Clear the slice
	*lastStatement = make([]token.Token, 0)
}

// AnalyzeArithmetic analyzes the given arithmetic expression
func AnalyzeArithmetic(
	statement []token.Token,
	variables *stack.Stack,
	functions *map[string]map[string]*code.Function,
	mods *map[string]map[string]*mod.FluentMod,
) {
	statementLen := len(statement)
	if statementLen == 0 {
		return
	}

	// Check if there is an arithmetic operator
	checkOperator(statement[0], statement)

	// Used to check if the next token is an operator
	expectingOperator := true

	lastStatement := make([]token.Token, 0)
	extractingIdentifier := false

	// Used to skip indexes
	skipToIndex := 0

	// The depth of parentheses
	parenDepth := 0

	for i, unit := range statement {
		if i < skipToIndex {
			continue
		}

		if extractingIdentifier {
			if isOperator(unit) {
				checkStatement(
					&extractingIdentifier,
					&lastStatement,
					unit,
					i,
					variables,
					functions,
					mods,
				)

				continue
			}

			lastStatement = append(lastStatement, unit)
			continue
		}

		if expectingOperator {
			if unit.GetType() == token.CloseParen {
				parenDepth--

				if parenDepth < 0 {
					logger.TokenError(
						unit,
						"Invalid arithmetic expression",
						"Unexpected closing parenthesis",
						"Check the arithmetic expression",
					)
				}

				continue
			}

			checkOperator(unit, statement[i:])
			expectingOperator = false

			continue
		}

		tokenType := unit.GetType()

		switch tokenType {
		case token.OpenParen:
			parenDepth++
			continue
		case token.CloseParen:
			// If this case is reached, it means
			// the code has something like: "()", which is invalid
			logger.TokenError(
				unit,
				"Invalid arithmetic expression",
				"An arithmetic expression must contain at least one number or variable",
			)
		case token.Identifier:
			extractingIdentifier = true
			lastStatement = append(lastStatement, unit)
			continue
		case token.DecimalLiteral, token.NumLiteral:
			expectingOperator = true
			continue
		default:
			logger.TokenError(
				unit,
				"Unexpected token",
				"Expected an identifier or a number",
				"Check the arithmetic expression",
			)
		}

	}

	// Check any remaining statement
	if len(lastStatement) > 0 {
		checkStatement(
			&extractingIdentifier,
			&lastStatement,
			statement[len(statement)-1],
			len(statement)-1,
			variables,
			functions,
			mods,
		)
	}
}
