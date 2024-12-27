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
	isLastValConstant := false

	// Used to check property access
	// i.e.: object.property
	lastValue := wrapper.NewFluentObject(dummyNothingType, nil)
	startAt := 0

	firstToken := statement[0]
	firstTokenType := firstToken.GetType()

	statementStr := ""

	for _, v := range statement {
		statementStr += v.GetValue()
	}

	beforeDot, _ := splitter.ExtractTokensBefore(
		statement,
		token.Dot,
		true,
		token.OpenParen,
		token.CloseParen,
		false,
	)

	switch firstTokenType {
	case token.New:
		AnalyzeObjectCreation(
			beforeDot,
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
		lastValue, isLastValConstant = converter.ToObj(firstToken, variables)
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

	switch remainingStatement[0].GetType() {
	case token.Dot:
		// Exclude the 1st character (either "." or "=")
		beforeAssignment, isAssignment := splitter.ExtractTokensBefore(
			remainingStatement[1:],
			token.Assign,
			false,
			token.Unknown,
			token.Unknown,
			false,
		)

		isFunCall = false

		if valueTypeWrapper.GetType() != types.ModType {
			logger.TokenError(
				remainingStatement[0],
				"Illegal property access",
				"Cannot access properties of a non-object",
				"Check the object type",
			)
		}

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
				functions,
				mods,
				&lastValue,
				&isFunCall,
				isAssignment,
			)
		}

		if isAssignment && isFunCall {
			logger.TokenError(
				remainingStatement[0],
				"Invalid statement",
				"A function call cannot be assigned to a property",
				"Check the statement",
			)
		} else if isAssignment {
			// PropertyAnalyzer checks for reassignments
			// of constant values, no need to check here
			afterAssignment := remainingStatement[len(beforeAssignment)+2:]

			AnalyzeType(
				afterAssignment,
				variables,
				functions,
				mods,
				lastValue,
			)
		}
	case token.Assign:
		if isLastValConstant {
			logger.TokenError(
				remainingStatement[0],
				"Invalid assignment",
				"Cannot assign to a constant value",
				"All literals also count as constant values",
				"Check if you are reassigning a literal",
			)
		}

		AnalyzeType(
			remainingStatement[1:],
			variables,
			functions,
			mods,
			lastValue,
		)
	default:
		logger.TokenError(
			remainingStatement[0],
			"Invalid statement",
			"This token was not expected at this position",
			"Check the statement",
		)
	}

	return lastValue
}
