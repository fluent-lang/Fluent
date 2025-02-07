/*
   The Fluent Programming Language
   -----------------------------------------------------
   Copyright (c) 2025 Rodrigo R. & All Fluent Contributors
   This program comes with ABSOLUTELY NO WARRANTY.
   For details type `fluent -l`. This is free software,
   and you are welcome to redistribute it under certain
   conditions; type `fluent -l -f` for details.
*/

package reassignment

import (
	error3 "fluent/analyzer/error"
	"fluent/analyzer/rule/expression"
	"fluent/analyzer/stack"
	"fluent/ast"
	"fluent/filecode"
	"fluent/filecode/types"
)

func AnalyzeReassignment(
	tree *ast.AST,
	variables *stack.ScopedStack,
	trace *filecode.FileCode,
) error3.Error {
	// Get the tree's children
	children := *tree.Children

	// Get the left expression
	leftExpr := children[0]
	rightExpr := children[1]

	// Analyze the property access
	obj, err := expression.AnalyzeExpression(
		leftExpr,
		trace,
		variables,
		false,
		&types.TypeWrapper{
			Children: &[]*types.TypeWrapper{},
		},
		true,
	)

	// Return the err if needed
	if err.Code != error3.Nothing {
		return err
	}

	// Define the expected type
	expected := obj.Type

	// Analyze the right expression
	obj, err = expression.AnalyzeExpression(
		rightExpr,
		trace,
		variables,
		false,
		&expected,
		false,
	)

	// Return the err if needed
	if err.Code != error3.Nothing {
		return err
	}

	return error3.Error{}
}
