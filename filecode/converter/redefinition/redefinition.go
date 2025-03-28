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

package redefinition

import (
	"fluent/filecode/function"
	"fluent/filecode/module"
	trace2 "fluent/filecode/trace"
	"fluent/logger"
	error2 "fluent/message/error"
	"fluent/util"
	"fmt"
	"os"
)

// CheckRedefinition checks if a given name is already defined in the provided map of defined values.
// If the name is already defined, it logs an error message and exits the program.
//
// Type Parameters:
//
//	T: A type that implements either the function.Function or module.Module interface.
//
// Parameters:
//
//	definedValues: A map of already defined values.
//	name: The name to check for redefinition.
//	entity: The entity being checked, which can be either a function or a module.
//	contents: The contents of the file where the entity is defined.
//	path: The path to the file where the entity is defined.
func CheckRedefinition[T function.Function | module.Module](
	definedValues map[string]*T,
	name string,
	entity any,
	contents string,
	path string,
) {
	var trace trace2.Trace

	if fn, ok := entity.(function.Function); ok {
		trace = fn.Trace
	} else if mod, ok := entity.(module.Module); ok {
		trace = mod.Trace
	} else {
		logger.Error("Unknown entity type")
		os.Exit(1)
	}

	if _, ok := definedValues[name]; ok {
		fmt.Print(error2.Redefinition(name))
		fmt.Print(
			util.BuildDetails(
				&contents,
				&path,
				trace.Line,
				trace.Column,
				true,
			),
		)
		logger.Info("'" + name + "' was previously defined here:")
		fmt.Print(
			util.BuildDetails(
				&contents,
				&path,
				trace.Line,
				trace.Column,
				true,
			),
		)

		os.Exit(1)
	}
}
