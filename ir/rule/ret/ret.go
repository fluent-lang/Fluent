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
	"fluent/ast"
	"fluent/filecode"
	"fluent/filecode/function"
	module2 "fluent/filecode/module"
	"fluent/filecode/types/wrapper"
	"fluent/ir/pool"
	"fluent/ir/rule/expression"
	"fluent/ir/tree"
	"fluent/ir/value"
	"fluent/ir/variable"
	"fluent/util"
	"strconv"
	"strings"
)

// MarshalReturn marshals a return statement into its string representation.
// Parameters:
// - representation: A pointer to a strings.Builder to store the marshaled return statement.
// - trace: A pointer to the FileCode structure containing the trace information.
// - fileCodeId: An integer representing the file code ID.
// - traceFileName: A string representing the name of the trace file.
// - isMod: A boolean indicating if the module is modified.
// - modulePropCounters: A pointer to a map of module property counters.
// - traceFn: A pointer to the Function structure representing the trace function.
// - originalPath: A pointer to a string representing the original path.
// - counter: A pointer to an integer counter.
// - element: A pointer to the AST structure representing the return statement.
// - variables: A pointer to a map of IRVariable structures representing the variables.
// - traceCounters: A pointer to the NumPool structure containing trace counters.
// - usedStrings: A pointer to the StringPool structure containing used strings.
// - localCounters: A pointer to a map of local counters.
// - retType: A pointer to the TypeWrapper structure representing the return type.
func MarshalReturn(
	representation *strings.Builder,
	trace *filecode.FileCode,
	fileCodeId int,
	traceFileName string,
	isMod bool,
	modulePropCounters *map[*module2.Module]*util.OrderedMap[string, *string],
	traceFn *function.Function,
	originalPath *string,
	counter *int,
	element *ast.AST,
	variables *map[string]*variable.IRVariable,
	traceCounters *pool.NumPool,
	usedStrings *pool.StringPool,
	localCounters *map[string]*string,
	retType *wrapper.TypeWrapper,
) {
	children := *element.Children

	// Check if this return in empty
	if len(children) == 0 {
		representation.WriteString("ret_void\n")
		return
	}

	// Get the returned expression
	expr := children[0]

	// Create a new instruction tree for this return
	retTree := tree.InstructionTree{
		Children:       nil,
		Representation: &strings.Builder{},
	}

	retTree.Representation.WriteString("ret ")

	// See if we can save memory in the expression
	if value.RetrieveStaticVal(fileCodeId, expr, retTree.Representation, usedStrings) {
		representation.WriteString(retTree.Representation.String())
		representation.WriteString("\n")
		return
	}

	retTree.Representation.WriteString("x")
	retTree.Representation.WriteString(strconv.Itoa(*counter))

	// Marshal the expression
	expression.MarshalExpression(
		representation,
		trace,
		traceFn,
		fileCodeId,
		isMod,
		traceFileName,
		originalPath,
		modulePropCounters,
		counter,
		expr,
		variables,
		traceCounters,
		usedStrings,
		localCounters,
		true,
		true,
		retType,
	)

	// Write the instruction tree to global tree
	representation.WriteString(retTree.Representation.String())
	representation.WriteString("\n")
}
