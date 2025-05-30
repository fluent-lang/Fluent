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

package expression

import (
	error3 "fluent/analyzer/error"
	"fluent/analyzer/object"
	queue2 "fluent/analyzer/queue"
	"fluent/analyzer/rule/arithmetic"
	"fluent/analyzer/rule/array"
	"fluent/analyzer/rule/boolean"
	"fluent/analyzer/rule/call"
	"fluent/analyzer/rule/property"
	"fluent/analyzer/stack"
	"fluent/ast"
	"fluent/filecode"
	"fluent/filecode/types/wrapper"
)

// literalRules is a map of rules that represent literals
var literalRules = map[ast.Rule]bool{
	ast.StringLiteral:        true,
	ast.NumberLiteral:        true,
	ast.BooleanLiteral:       true,
	ast.DecimalLiteral:       true,
	ast.Array:                true,
	ast.ArithmeticExpression: true,
	ast.BooleanExpression:    true,
}

// AnalyzeExpression analyzes an AST expression and returns the resulting object and any errors encountered.
//
// Parameters:
// - tree: The AST tree representing the expression.
// - trace: The file code trace for debugging and error reporting.
// - variables: The stack of scoped variables available in the current context.
// - enforceHeapRequirement: A boolean indicating whether heap allocation requirements should be enforced.
// - firstExpected: The expected type of the expression.
// - isPropReassignment: A boolean indicating whether the current element comes from a property reassignment.
// - allowPointers: A boolean indicating whether pointers are allowed.
//
// Returns:
// - object.Object: The resulting object after analyzing the expression.
// - error3.Error: Any error encountered during the analysis.
func AnalyzeExpression(
	tree *ast.AST,
	trace *filecode.FileCode,
	variables *stack.ScopedStack,
	enforceHeapRequirement bool,
	firstExpected *wrapper.TypeWrapper,
	isPropReassignment bool,
	allowPointers bool,
	allowedIds []int,
) (*object.Object, *error3.Error) {
	// Set the inferred type of the tree if we hav ea first expected element
	if firstExpected != nil {
		tree.InferredType = firstExpected
	}

	result := object.Object{
		Type: wrapper.TypeWrapper{
			Children: &[]*wrapper.TypeWrapper{},
		},
	}

	// Use a queue to analyze the expression
	queue := []queue2.ExpectedPair{
		{
			Expected:           firstExpected,
			Got:                &result,
			Tree:               tree,
			HasMetDereference:  false,
			ActualPointers:     0,
			IsParam:            allowPointers,
			IsPropReassignment: isPropReassignment,
		},
	}

	// Use the queue
	for len(queue) > 0 {
		// Pop the first element
		element := queue[0]
		queue = queue[1:]

		// Used to skip nodes
		startAt := 0

		// Used to keep track of whether the current value
		// has nested expressions
		hasNested := false

		for _, node := range *element.Tree.Children {
			hasToBreak := false

			switch node.Rule {
			case ast.Pointer:
				startAt++
				// Increment the pointer count
				element.Got.Type.PointerCount++
				element.HasPointers = true

				if element.HasMetDereference {
					element.ActualPointers++
				}
			case ast.Dereference:
				startAt++
				element.HasMetDereference = true

				// Decrement the pointer count
				element.ActualPointers--
				element.Got.Type.PointerCount--
			default:
				hasToBreak = true
			}

			if hasToBreak {
				break
			}
		}

		// Check for illegal pointers
		if element.HasPointers && !element.IsParam {
			return nil, &error3.Error{
				Code:   error3.InvalidPointer,
				Line:   element.Tree.Line,
				Column: element.Tree.Column,
			}
		}

		// Get the child
		child := (*element.Tree.Children)[startAt]

		// See if the address of this value can be taken
		if literalRules[child.Rule] && element.Got.Type.PointerCount > 0 {
			return nil, &error3.Error{
				Code:   error3.CannotTakeAddress,
				Line:   child.Line,
				Column: child.Column,
			}
		}

		switch child.Rule {
		case ast.StringLiteral:
			element.Got.Type.BaseType = "str"
			element.Got.Type.IsPrimitive = true
		case ast.NumberLiteral:
			element.Got.Type.BaseType = "num"
			element.Got.Type.IsPrimitive = true
		case ast.BooleanLiteral:
			element.Got.Type.BaseType = "bool"
			element.Got.Type.IsPrimitive = true
		case ast.DecimalLiteral:
			element.Got.Type.BaseType = "dec"
			element.Got.Type.IsPrimitive = true
		case ast.Identifier:
			// Check for property access
			if element.IsPropAccess {
				err := property.ProcessPropIdentifier(
					&element,
					trace,
					child,
				)

				// Return the error if it is not nothing
				if err != nil {
					return nil, err
				}
			} else {
				// Check if the variable exists
				value := variables.Load(child.Value, allowedIds)

				if value == nil {
					return nil, &error3.Error{
						Code:       error3.UndefinedReference,
						Additional: []string{*child.Value},
						Line:       element.Tree.Line,
						Column:     element.Tree.Column,
					}
				}

				// Check for reassignments
				if element.IsPropReassignment && value.Constant {
					return nil, &error3.Error{
						Code:   error3.ConstantReassignment,
						Line:   element.Tree.Line,
						Column: element.Tree.Column,
					}
				}

				oldPointerCount := element.Got.Type.PointerCount
				element.Got.Type = value.Value.Type
				element.Got.Type.PointerCount += oldPointerCount
				element.ActualPointers += value.Value.Type.PointerCount
				element.Got.Value = value.Value.Value
				element.Got.IsHeap = value.Value.IsHeap
			}
		case ast.Array:
			err := array.AnalyzeArray(child, element.Expected, &queue)

			// Return the error if it is not nothing
			if err != nil {
				return nil, err
			}

			element.Got.Type = *element.Expected
		case ast.FunctionCall, ast.ObjectCreation:
			// This will later be fully determined by the call analyzer
			element.Got.IsHeap = false

			// Pass the input to the function call analyzer
			err := call.AnalyzeFunctionCall(
				child,
				trace,
				&element,
				&queue,
				child.Rule == ast.ObjectCreation,
			)

			// Return the error if it is not nothing
			if err != nil {
				return nil, err
			}
		case ast.Expression:
			hasNested = true

			// Add the expression to the queue
			queue = append([]queue2.ExpectedPair{{
				Expected:          element.Expected,
				Got:               element.Got,
				Tree:              child,
				HasMetDereference: element.HasMetDereference,
				ActualPointers:    element.ActualPointers,
				IsArithmetic:      element.IsArithmetic,
				IsParam:           element.IsParam,
			}}, queue...)

			element.Got.Type = *element.Expected
		case ast.PropertyAccess:
			// This will later be fully determined by the property access analyzer
			element.Got.IsHeap = false

			// Pass the input to the property access analyzer
			property.AnalyzePropertyAccess(
				&element,
				child,
				&queue,
				isPropReassignment,
			)

			hasNested = true
		case ast.ArithmeticExpression:
			// Pass the input to the arithmetic analyzer
			err := arithmetic.AnalyzeArithmetic(
				child,
				&element,
				&queue,
			)

			// Return the error if it is not nothing
			if err != nil {
				return nil, err
			}
		case ast.BooleanExpression:
			// Pass the input to the boolean analyzer
			boolean.AnalyzeBoolean(
				child,
				&element,
				&queue,
			)
		default:
		}

		// isInferred does not work here because it was defined
		// before the switch statement
		if element.Expected.BaseType == "(Infer)" {
			oldPointerCount := element.Expected.PointerCount
			oldArrayCount := element.Expected.ArrayCount

			*element.Expected = element.Got.Type
			element.Expected.PointerCount = oldPointerCount
			element.Expected.ArrayCount = oldArrayCount
		}

		// Check if the pointer count is negative
		if !hasNested && element.ActualPointers < 0 {
			return nil, &error3.Error{
				Code:   error3.InvalidDereference,
				Line:   element.Tree.Line,
				Column: element.Tree.Column,
			}
		}

		if hasNested {
			continue
		}

		// Check for type mismatch
		if element.Expected.BaseType != "" && !element.Expected.Compare(element.Got.Type) {
			return nil, &error3.Error{
				Code:       error3.TypeMismatch,
				Line:       element.Tree.Line,
				Column:     element.Tree.Column,
				Additional: []string{element.Expected.Marshal(), element.Got.Type.Marshal()},
			}
		}

		// Check if the data escapes the function
		if enforceHeapRequirement && element.HeapRequired && !element.Got.IsHeap {
			return nil, &error3.Error{
				Code:   error3.DataOutlivesStack,
				Line:   element.Tree.Line,
				Column: element.Tree.Column,
			}
		}

		if element.ModRequired && element.Got.Value == nil {
			return nil, &error3.Error{
				Code:       error3.TypeMismatch,
				Line:       element.Tree.Line,
				Column:     element.Tree.Column,
				Additional: []string{"Module", element.Got.Type.Marshal()},
			}
		}

		// Check for arithmetic operations
		if element.IsArithmetic && element.Got.Type.BaseType != "num" && element.Got.Type.BaseType != "dec" && element.Got.Type.BaseType != "(Infer)" {
			return nil, &error3.Error{
				Code:       error3.TypeMismatch,
				Line:       element.Tree.Line,
				Column:     element.Tree.Column,
				Additional: []string{"num or dec", element.Got.Type.Marshal()},
			}
		}

		// Set the inferred type
		if element.Got.Type.BaseType != "" && element.Tree.InferredType == nil {
			element.Tree.InferredType = &element.Got.Type
		}
	}

	return &result, nil
}
