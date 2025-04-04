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

package property

import (
	"fluent/analyzer/queue"
	"fluent/filecode/module"
)

// EvaluateLastPropValue evaluates the last property value of the given element.
// If the element is a property access and the last property value is not nil,
// it attempts to cast the value to a module.Module. If the cast is successful,
// it returns the module. Otherwise, it returns nil.
//
// Parameters:
//   - element: A pointer to a queue.ExpectedPair which contains the property access information.
//
// Returns:
//   - A pointer to a module.Module if the cast is successful.
func EvaluateLastPropValue(element *queue.ExpectedPair) *module.Module {
	if element.IsPropAccess {
		if element.LastPropValue == nil {
			return nil
		}

		var convert interface{}
		convert = *element.LastPropValue

		// Cast the last property value to a module
		mod, castOk := convert.(module.Module)

		if !castOk {
			nMod, nCastOk := convert.(*module.Module)

			if nCastOk {
				castOk = true
				mod = *nMod
			}
		}

		if !castOk {
			return nil
		}

		return &mod
	}

	return nil
}
