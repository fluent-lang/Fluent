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
	"fluent/ast"
	"fluent/filecode"
	"fluent/ir/pool"
	"fluent/ir/rule/call"
	"fluent/ir/tree"
	"strconv"
	"strings"
)

func MarshalExpression(
	funTree *tree.InstructionTree,
	trace *filecode.FileCode,
	traceFileName string,
	counter *pool.CounterPool,
	element *ast.AST,
	traceMagicCounter *int,
	variables map[string]string,
	traceCounters *map[int]string,
	usedStrings *pool.StringPool,
	nameCounters *map[string]map[string]string,
	localCounters map[string]string,
) {
	result := tree.InstructionTree{
		Children:       &[]*tree.InstructionTree{},
		Representation: &strings.Builder{},
	}

	// Use a queue to process the elements of the AST
	queue := []tree.MarshalPair{
		{
			Child:    element,
			Parent:   &result,
			IsInline: false,
		},
	}

	for len(queue) > 0 {
		// Get the first element of the queue
		pair := queue[0]
		queue = queue[1:]

		// Get the children of the current element
		children := *pair.Child.Children
		// Used to skip pointers and dereferences
		startAt := 0

		// Move values to the stack for parameters
		if pair.IsParam {
			pair.Parent.Representation.WriteString("mov x")
			pair.Parent.Representation.WriteString(strconv.Itoa(pair.Counter))
			pair.Parent.Representation.WriteString(" ")
			pair.Parent.Representation.WriteString(pair.Expected.Marshal())
			pair.Parent.Representation.WriteString(" ")
		}

		// Add pointers and dereferences
		for _, child := range children {
			if child.Rule == ast.Pointer {
				pair.Parent.Representation.WriteString("&")
				startAt++
			} else if child.Rule == ast.Dereference {
				pair.Parent.Representation.WriteString("*")
				startAt++
			} else {
				break
			}
		}

		// Get the remaining expression
		child := children[startAt]

		switch child.Rule {
		case ast.FunctionCall:
			call.MarshalFunctionCall(
				&result,
				child,
				traceFileName,
				trace,
				traceMagicCounter,
				counter,
				pair.Parent,
				traceCounters,
				nameCounters,
				usedStrings,
				&queue,
			)
		case ast.Identifier:
			// Retrieve the variable
			variable := variables[*child.Value]
			pair.Parent.Representation.WriteString(variable)
		case ast.StringLiteral:
			// Request an address space for the string literal
			pair.Parent.Representation.WriteString(usedStrings.RequestAddress(*child.Value))
		case ast.NumberLiteral, ast.DecimalLiteral:
			// Directly write the tree's value
			pair.Parent.Representation.WriteString(*child.Value)
		case ast.BooleanLiteral:
			// Write 1 if the value is true, 0 otherwise
			var val string
			if *child.Value == "true" {
				val = "1"
			} else {
				val = "0"
			}
			pair.Parent.Representation.WriteString(val)
		default:
		}
	}

	// Append all children to the parent tree
	for _, child := range *result.Children {
		funTree.Representation.WriteString(child.Representation.String())
		funTree.Representation.WriteString("\n")
	}

	// Append the expression itself (without the children)
	funTree.Representation.WriteString(result.Representation.String())
	funTree.Representation.WriteString("\n")
}
