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

package converter

import (
	"fluent/ansi"
	ast2 "fluent/ast"
	"fluent/filecode"
	"fluent/filecode/converter/function"
	module2 "fluent/filecode/converter/module"
	"fluent/filecode/converter/redefinition"
	function2 "fluent/filecode/function"
	"fluent/filecode/module"
	"fluent/lexer"
	"fluent/logger"
	"fluent/parser"
	"fluent/state"
	"fluent/util"
	"fmt"
	"os"
	"strings"
)

var StdPath = os.Getenv("FLUENT_STD_PATH")

// The system's path separator
const pathSeparator = string(os.PathSeparator)

type queueElement struct {
	path     *string
	trace    *ast2.AST
	contents *string
}

// ConvertToFileCode converts the given entry file and all its imports to a map of FileCode.
// It checks for the existence of the standard library path, reads files, tokenizes them, parses them into ASTs,
// and then converts the ASTs into FileCode structures. It also handles circular import detection.
//
// Parameters:
// - entry: The entry file path to start the conversion.
// - silent: Whether to suppress the output.
//
// Returns:
// - A map where the keys are file paths and the values are FileCode structures.
func ConvertToFileCode(entry string, silent bool) map[string]filecode.FileCode {
	// Check that the stdlib path exists
	if StdPath == "" || !util.DirExists(StdPath) {
		logger.Error("The FLUENT_STD_PATH environment variable is not set")
		logger.Help("Try reinstalling the CLI")
		os.Exit(1)
	}

	// Use a queue to convert the file and all of its imports
	queue := []queueElement{
		{
			path: &entry,
		},
	}

	// Save the seen imports in a map (for O(1) lookup) to detect circular imports
	seenImports := map[string]bool{}
	// Save a slice of the seen imports to print the full import chain in a sorted manner
	var seenImportsSlice []string
	result := make(map[string]filecode.FileCode)

	for len(queue) > 0 {
		// Get the first element of the queue
		element := queue[0]
		queue = queue[1:]

		path := element.path
		trace := element.trace
		elContents := element.contents

		isStd := strings.HasPrefix(*path, StdPath)

		// Save the already-processed std imports
		processedStdImports := map[string]bool{}

		// Detect circular imports
		if seenImports[*path] {
			logger.Error("Circular import detected")
			logger.Info("Full import chain:")

			spaces := 0
			// Use a builder to print the full import chain in an efficient manner
			builder := strings.Builder{}

			for _, importPath := range seenImportsSlice {
				if *path == importPath {
					builder.WriteString(
						logger.BuildInfo(
							ansi.Colorize(
								ansi.BoldBrightRed,
								strings.Repeat("  ", spaces)+"-> "+util.DiscardCwd(importPath)+" (Circular)",
							),
						),
					)
				} else {
					builder.WriteString(
						logger.BuildInfo(strings.Repeat("  ", spaces) + "-> " + util.DiscardCwd(importPath)),
					)
				}

				spaces++
			}
			fmt.Print(builder.String())

			// Also print the current circular import's details
			logger.Info(
				ansi.Colorize(
					ansi.BoldBrightRed,
					strings.Repeat("  ", spaces)+"-> "+util.DiscardCwd(*path)+" (Circular)",
				),
			)

			logger.Info("Full details:")
			fmt.Print(util.BuildDetails(elContents, path, trace.Line, trace.Column, true))

			os.Exit(1)
		}

		if !isStd {
			seenImports[*path] = true
			seenImportsSlice = append(seenImportsSlice, *path)
		}

		// Read the file
		contents := util.ReadFile(*path)
		fileName := util.FileName(path)

		// Lex the file
		if !silent {
			state.Emit(state.Lexing, fileName)
		}

		tokens, lexerError := lexer.Lex(contents, *path)

		if lexerError != nil {
			state.FailAllSpinners()
			// Build and print the error
			util.PrintError(&contents, path, &lexerError.Message, lexerError.Line, lexerError.Column)
			os.Exit(1)
		}

		state.PassAllSpinners()
		if !silent {
			state.Emit(state.Parsing, fileName)
		}

		// Parse the tokens to an AST
		ast, parsingError := parser.Parse(tokens, *path)

		if parsingError != nil {
			state.FailAllSpinners()
			errorMessage := util.BuildMessageFromParsingError(*parsingError)

			// Build and print the error
			util.PrintError(&contents, path, &errorMessage, parsingError.Line, parsingError.Column)
			os.Exit(1)
		}

		state.PassAllSpinners()

		if !silent {
			state.Emit(state.Processing, fileName)
		}

		code := filecode.FileCode{
			Path:      *path,
			Functions: make(map[string]*function2.Function),
			Modules:   make(map[string]*module.Module),
			Imports:   make([]string, 0),
			Contents:  contents,
		}

		// Traverse the AST to convert it to a FileCode
		for _, child := range *ast.Children {
			rule := child.Rule

			switch rule {
			case ast2.Import:
				// Get the path
				importPath := *(*child.Children)[0].Value

				if !strings.HasSuffix(importPath, ".fluent") {
					importPath += ".fluent"
				}

				isStd := strings.HasPrefix(importPath, "@std")

				if isStd {
					importPath = strings.Replace(importPath, "@std", StdPath, 1)
					importPath = strings.ReplaceAll(importPath, "::", pathSeparator)

					if processedStdImports[importPath] {
						continue
					}

					processedStdImports[importPath] = true
				} else {
					// Get the file's directory
					dir := util.GetDir(*path)

					// Join the directory with the path
					importPath = dir + pathSeparator + importPath
				}

				// Append the path to the code's imports
				code.Imports = append(code.Imports, importPath)

				// Queue the path
				queue = append(queue, queueElement{
					path:     &importPath,
					contents: &contents,
					trace:    child,
				})
			case ast2.Function:
				// Convert to a Function wrapper
				fn := function.ConvertFunction(child, isStd)

				// Check for redefinitions
				redefinition.CheckRedefinition(code.Functions, fn.Name, fn, contents, *path)

				code.Functions[fn.Name] = &fn
			case ast2.Module:
				// Convert to a Function wrapper
				mod := module2.ConvertModule(child, contents)

				// Check for redefinitions
				redefinition.CheckRedefinition(code.Modules, mod.Name, mod, contents, *path)

				code.Modules[mod.Name] = &mod
			default:
			}
		}

		state.PassAllSpinners()
		result[*path] = code
	}

	return result
}
