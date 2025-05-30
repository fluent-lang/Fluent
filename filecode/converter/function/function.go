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

package function

import (
	ast2 "fluent/ast"
	"fluent/filecode/function"
	"fluent/filecode/trace"
	"fluent/filecode/types"
)

// ConvertFunction converts an AST node to a function representation.
// Parameters:
//   - ast: A pointer to the AST node to be converted.
//   - isStd: A boolean indicating whether the function is a standard library function.
//
// Returns:
//   - A function.Function representing the converted AST node.
func ConvertFunction(ast *ast2.AST, isStd bool) function.Function {
	result := function.Function{
		Params: make([]function.Param, 0),
		Path:   *ast.File,
		Trace: trace.Trace{
			Line:   ast.Line,
			Column: ast.Column,
		},
	}

	children := *ast.Children
	var block *ast2.AST
	var public bool
	var returnType *ast2.AST
	var templates *ast2.AST
	var params *ast2.AST
	var name string

	// Parse the function's properties
	for _, child := range children {
		switch child.Rule {
		case ast2.Public:
			public = true
		case ast2.Identifier:
			name = *child.Value
		case ast2.Parameters:
			params = child
		case ast2.Type:
			returnType = child
		case ast2.Block:
			block = child
		case ast2.Templates:
			templates = child
		default:
		}
	}

	if returnType == nil {
		returnType = &ast2.AST{
			Line:   ast.Line,
			Column: ast.Column,
			Rule:   ast2.Type,
			Children: &[]*ast2.AST{
				{
					Line:   ast.Line,
					Column: ast.Column,
					Rule:   ast2.Nothing,
					Value:  nil,
				},
			},
		}
	}

	if block == nil {
		// -- Impossible case --
		return function.Function{}
	}

	// Set the function's properties
	result.Public = public
	result.Name = name
	result.Body = *block
	result.ReturnType = types.ConvertToTypeWrapper(*returnType)
	result.Templates = make(map[string]bool)
	result.IsStd = isStd

	// Handle templates
	if templates != nil {
		for _, template := range *templates.Children {
			result.Templates[*template.Value] = true
		}
	}

	// Parse the parameters
	funParams := make([]function.Param, 0)

	if params != nil {
		for _, param := range *params.Children {
			paramName := (*param.Children)[0].Value
			paramType := types.ConvertToTypeWrapper(*(*param.Children)[1])

			// Add the parameter to the function's parameters
			funParams = append(funParams, function.Param{
				Type: paramType,
				Trace: trace.Trace{
					Line:   param.Line,
					Column: param.Column,
				},
				Name: *paramName,
			})
		}
	}

	// Set the function's parameters
	result.Params = funParams

	return result
}
