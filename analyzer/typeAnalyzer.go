package analyzer

import (
	"fluent/code"
	"fluent/code/mod"
	"fluent/code/types"
	"fluent/code/wrapper"
	"fluent/logger"
	"fluent/stack"
	"fluent/token"
)

// AnalyzeType analyzes the given type and makes sure it matches the expected type
func AnalyzeType(
	statement []token.Token,
	variables *stack.Stack,
	functions *map[string]map[string]*code.Function,
	mods *map[string]map[string]*mod.FluentMod,
	expected wrapper.FluentObject,
	enforceGenericsMatch bool,
) {
	expectedTypeWrapper := expected.GetType()
	isMod := expectedTypeWrapper.GetType() == types.ModType

	// Analyze the type
	value := AnalyzeStatement(
		statement,
		variables,
		functions,
		mods,
		expectedTypeWrapper,
	)

	if isMod {
		valueTypeWrapper := value.GetType()

		if valueTypeWrapper.GetType() != types.ModType {
			logger.TokenError(
				statement[0],
				"Type mismatch",
				"This type does not match the value type",
				"Change the declaration or remove the assignment",
				"Expected: "+expectedTypeWrapper.Marshal(),
				"Got: "+valueTypeWrapper.Marshal(),
			)
		}

		if !expectedTypeWrapper.Compare(valueTypeWrapper) {

			if !enforceGenericsMatch && expectedTypeWrapper.GetBaseType() == valueTypeWrapper.GetBaseType() {
				return
			}

			logger.TokenError(
				statement[0],
				"Type mismatch",
				"This type does not match the value type",
				"Change the declaration or remove the assignment",
				"Expected: "+expectedTypeWrapper.Marshal(),
				"Got: "+valueTypeWrapper.Marshal(),
			)
		}
	} else if !expectedTypeWrapper.Compare(value.GetType()) {
		valueType := value.GetType()
		logger.TokenError(
			statement[0],
			"Type mismatch",
			"This type does not match the value type",
			"Change the declaration or remove the assignment",
			"Expected: "+expectedTypeWrapper.Marshal(),
			"Got: "+valueType.Marshal(),
		)
	}
}
