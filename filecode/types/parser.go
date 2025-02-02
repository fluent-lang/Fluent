/*
   The Fluent Programming Language
   -----------------------------------------------------
   Copyright (c) 2025 Rodrigo R. & All Fluent Contributors
   This program comes with ABSOLUTELY NO WARRANTY.
   For details type `fluent -l`. This is free software,
   and you are welcome to redistribute it under certain
   conditions; type `fluent -l -f` for details.
*/

package types

import (
	"fluent/ast"
	"fluent/filecode/trace"
)

type queueElement struct {
	node   ast.AST
	parent *TypeWrapper
}

// ConvertToTypeWrapper converts an AST to a TypeWrapper.
// It processes the AST in a breadth-first manner using a queue.
// Parameters:
// - tree: the root of the AST to be converted.
// Returns:
// - A TypeWrapper representing the structure of the AST.
func ConvertToTypeWrapper(tree ast.AST) TypeWrapper {
	result := TypeWrapper{
		Trace: trace.Trace{
			Line:   tree.Line,
			Column: tree.Column,
		},
		Children: &[]*TypeWrapper{},
	}

	// Use a queue to process the AST in a breadth-first manner.
	queue := []queueElement{
		{
			node:   tree,
			parent: &result,
		},
	}

	for len(queue) > 0 {
		// Get the first element of the queue
		element := queue[0]
		queue = queue[1:]

		// Get the node and its parent
		node := element.node
		parent := element.parent

		// Iterate over the node's children
		for _, child := range *node.Children {
			rule := child.Rule

			switch rule {
			case ast.Pointer:
				// Increment the pointer count
				parent.PointerCount++
			case ast.ArrayType:
				// Increment the array count
				parent.ArrayCount++
			case ast.Identifier, ast.Primitive:
				parent.BaseType = *child.Value
				parent.IsPrimitive = rule == ast.Primitive
			case ast.Type:
				// Create a new TypeWrapper for the child
				newType := TypeWrapper{
					Trace: trace.Trace{
						Line:   child.Line,
						Column: child.Column,
					},
					Children: &[]*TypeWrapper{},
				}

				// Add the new TypeWrapper to the parent's children
				*parent.Children = append(*parent.Children, &newType)

				// Add the new TypeWrapper to the queue
				queue = append(queue, queueElement{
					node:   *child,
					parent: &newType,
				})
			default:
			}
		}
	}

	return result
}
