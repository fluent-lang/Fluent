package analyzer

import (
	"fluent/code"
	"fluent/code/mod"
	"fluent/code/types"
	"fluent/code/wrapper"
	"fluent/logger"
	"fluent/stack"
	"fluent/token"
	"fluent/tokenUtil/converter"
	"fluent/tokenUtil/splitter"
)

// AnalyzeStatement analyzes the given statement
// and returns the object type that the statement
// returns
func AnalyzeStatement(
	statement []token.Token,
	variables *stack.Stack,
	functions *map[string]map[string]*code.Function,
	mods *map[string]map[string]*mod.FluentMod,
	inferToType wrapper.TypeWrapper,
) wrapper.FluentObject {
	// Used to know what to check for
	isArithmetic := false
	isFunCall := false

	// Used to check property access
	// i.e.: object.property
	lastValue := wrapper.NewFluentObject(dummyNothingType, nil)
	startAt := 0

	firstToken := statement[0]
	firstTokenType := firstToken.GetType()

	switch firstTokenType {
	case token.New:
		AnalyzeObjectCreation(
			splitter.ExtractTokensBefore(
				statement,
				token.Dot,
				true,
				token.OpenParen,
				token.CloseParen,
				false,
			),
			variables,
			functions,
			mods,
			&startAt,
			&lastValue,
			inferToType,
		)

		break
	case token.Identifier:
		AnalyzeIdentifier(
			statement,
			variables,
			functions,
			mods,
			&startAt,
			&lastValue,
			&isArithmetic,
			&isFunCall,
		)

		break
	default:
		lastValue = converter.ToObj(firstToken, variables)
		valueTypeWrapper := lastValue.GetType()

		isArithmetic = valueTypeWrapper.GetType() == types.IntType || valueTypeWrapper.GetType() == types.DecimalType
		startAt = 1
	}

	// Analyze the rest of the statement
	remainingStatement := statement[startAt:]

	if len(remainingStatement) == 0 {
		return lastValue
	}

	if isArithmetic {
		AnalyzeArithmetic(
			remainingStatement,
			variables,
			functions,
			mods,
		)

		return lastValue
	}

	valueTypeWrapper := lastValue.GetType()
	if valueTypeWrapper.GetType() != types.ModType {
		logger.TokenError(
			remainingStatement[0],
			"Illegal property access",
			"Cannot access properties of a non-object",
			"Check the object type",
		)
	}

	// Analyze reassignments
	if remainingStatement[0].GetType() == token.Assign {
		if isFunCall {
			logger.TokenError(
				remainingStatement[0],
				"Invalid operation",
				"Cannot assign to a method call",
				"Check the statement",
			)
		}

		variable, _ := variables.Load(firstToken.GetValue())
		// No need to check if it was found here, it was already checked

		if variable.IsConstant() {
			logger.TokenError(
				remainingStatement[0],
				"Cannot reassign constant",
				"Check the variable declaration",
			)
		}

		return lastValue
	}

	// The only valid operation after all that has been processed
	// is property access, therefore the fist token of the remaining
	// statement must be a dot
	if remainingStatement[0].GetType() != token.Dot {
		logger.TokenError(
			remainingStatement[0],
			"Invalid operation",
			"Invalid operation after identifier",
			"Check the statement",
		)
	}

	// Get tokens before an assignment
	// i.e.: object.property = value
	beforeAssignment := splitter.ExtractTokensBefore(
		remainingStatement[1:],
		token.Assign,
		false,
		token.Unknown,
		token.Unknown,
		false,
	)

	// if beforeAssignment is empty, that means that
	// the statement ends in a dot: "object.property."
	// which is invalid
	if len(beforeAssignment) == 0 {
		logger.TokenError(
			remainingStatement[0],
			"Invalid operation",
			"Invalid operation after identifier",
			"Check the statement",
		)
	}

	// +1 for the dot
	// +1 for the assignment
	afterAssignment := remainingStatement[len(beforeAssignment)+2:]

	// Reset isFunCall to catch assignments to methods
	isFunCall = false
	props := splitter.SplitTokens(
		beforeAssignment,
		token.Dot,
		token.OpenParen,
		token.CloseParen,
	)

	// Analyze all props
	for _, prop := range props {
		AnalyzePropAccess(
			prop,
			variables,
			functions,
			mods,
			&lastValue,
			&isFunCall,
			len(afterAssignment) > 0,
		)
	}

	if isFunCall && len(afterAssignment) > 0 {
		logger.TokenError(
			afterAssignment[0],
			"Invalid operation",
			"Cannot assign to a method call",
			"Check the statement",
		)
	}

	// Analyze assignment
	AnalyzeType(
		afterAssignment,
		variables,
		functions,
		mods,
		lastValue,
		true,
	)

	return lastValue
}
