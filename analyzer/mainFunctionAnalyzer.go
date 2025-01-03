package analyzer

import (
	"fluent/code"
	"fluent/code/types"
	"fluent/logger"
)

// AnalyzeMainFunc analyzes the main function
func AnalyzeMainFunc(function *code.Function) {
	// Main function can't be public
	// Technically, it can be, but it's better
	// to enforce this rule because another function
	// calling the main function will cause
	// an infinite loop
	if function.IsPublic() {
		logger.TokenError(
			function.GetTrace(),
			"Main function cannot be public",
			"Remove the 'pub' keyword from the main function",
		)
	}

	if len(function.GetParameters()) != 0 {
		logger.TokenError(
			function.GetTrace(),
			"Main function cannot have parameters",
			"Remove the parameters from the main function",
		)
	}

	returnType := function.GetReturnType()

	if returnType.GetType() != types.NothingType {
		logger.TokenError(
			function.GetTrace(),
			"Main function must return nothing",
			"Change the function's return type to nothing",
		)
	}
}
