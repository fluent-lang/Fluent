/*
   The Fluent Programming Language
   -----------------------------------------------------
   This code is released under the GNU GPL v3 license.
   For more information, please visit:
   https://www.gnu.org/licenses/gpl-3.0.html
   -----------------------------------------------------
   Copyright (c) 2025 Rodrigo R. & All Fluent Contributors
   This program comes with ABSOLUTELY NO WARRANTY.
   For details type `fluent l`. This is free software,
   and you are welcome to redistribute it under certain
   conditions; type `fluent l -f` for details.
*/

package ret

import (
	error3 "fluent/analyzer/error"
	"fluent/analyzer/rule/expression"
	"fluent/analyzer/stack"
	"fluent/ast"
	"fluent/filecode"
	"fluent/filecode/types/wrapper"
)

// AnalyzeReturn analyzes the return statement of a function.
// It checks if the return statement matches the expected return type,
// if the data escapes the function, and if there are any type mismatches.
//
// Parameters:
// - tree: the AST of the return statement
// - trace: the file code trace
// - variables: the scoped stack of variables
// - expected: the expected return type
// - allowedIds: A slice of allowed IDs that the current block chain holds.
//
// Returns:
// - error3.Error: an error object indicating any issues found during analysis
func AnalyzeReturn(
	tree *ast.AST,
	trace *filecode.FileCode,
	variables *stack.ScopedStack,
	expected *wrapper.TypeWrapper,
	allowedIds []int,
) *error3.Error {
	// Check if the tree has children
	if len(*tree.Children) == 0 {
		return nil
	}

	// Check if the function doesn't expect a return value
	exprNode := (*tree.Children)[0]
	if expected.BaseType == "nothing" {
		return &error3.Error{
			Code:   error3.ShouldNotReturn,
			Line:   exprNode.Line,
			Column: exprNode.Column,
		}
	}

	// Analyze the expression
	expr, err := expression.AnalyzeExpression(
		exprNode,
		trace,
		variables,
		true,
		expected,
		false,
		true,
		allowedIds,
	)

	// Return the error if there is one
	if err != nil {
		return err
	}

	// Check if the data escapes the function
	if !expr.IsHeap && expr.Type.PointerCount > 0 {
		return &error3.Error{
			Code:   error3.DataOutlivesStack,
			Line:   exprNode.Line,
			Column: exprNode.Column,
		}
	}

	return nil
}
