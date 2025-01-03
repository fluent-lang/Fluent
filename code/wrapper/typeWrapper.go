package wrapper

import (
	"fluent/code/types"
	"fluent/logger"
	"fluent/token"
	"fluent/tokenUtil/inferrer"
	"fluent/tokenUtil/splitter"
	"strings"
)

// TypeWrapper is a wrapper for a data type
// based on the given tokens
// For example: Result<str, bool>
type TypeWrapper struct {
	baseType   string                 // Result
	parameters []TypeWrapper          // [str, bool]
	objType    types.FluentObjectType // The object type
}

// ForceNewTypeWrapper creates a new TypeWrapper
// without checking if the data type is valid
func ForceNewTypeWrapper(
	baseType string,
	parameters []TypeWrapper,
	objType types.FluentObjectType,
) TypeWrapper {
	return TypeWrapper{
		baseType:   baseType,
		parameters: parameters,
		objType:    objType,
	}
}

// NewTypeWrapper creates a new TypeWrapper
func NewTypeWrapper(
	tokens []token.Token,
	trace token.Token,
) TypeWrapper {
	// Parse the tokens
	if len(tokens) == 0 {
		logger.TokenError(
			trace,
			"Invalid data type",
			"This data type is empty",
		)
	}

	baseType := tokens[0]
	parameters := make([]TypeWrapper, 0)

	// Check if the data type has parameters
	if len(tokens) > 1 {
		if tokens[1].GetType() != token.LessThan {
			logger.TokenError(
				tokens[1],
				"Invalid data type",
				"Expected '<' after the base type",
			)
		}

		if tokens[len(tokens)-1].GetType() != token.GreaterThan {
			logger.TokenError(
				tokens[len(tokens)-1],
				"Invalid data type",
				"Expected '>' at the end of the data type",
			)
		}

		// Split by commas what's inside the '<' and '>'
		// For example: Result<str, bool>
		paramsTokens := splitter.SplitTokens(
			tokens[2:len(tokens)-1],
			token.Comma,
			token.LessThan,
			token.GreaterThan,
		)

		for _, paramTokens := range paramsTokens {
			parameters = append(parameters, NewTypeWrapper(paramTokens, trace))
		}
	}

	return TypeWrapper{
		baseType:   baseType.GetValue(),
		parameters: parameters,
		objType:    inferrer.InferFromRawType(baseType),
	}
}

// GetBaseType returns the type of the base object
func (tw *TypeWrapper) GetBaseType() string {
	return tw.baseType
}

// GetParameters returns the parameters of the data type
func (tw *TypeWrapper) GetParameters() []TypeWrapper {
	return tw.parameters
}

// GetType returns the object type
func (tw *TypeWrapper) GetType() types.FluentObjectType {
	return tw.objType
}

// Compare compares two TypeWrappers
func (tw *TypeWrapper) Compare(other TypeWrapper) bool {
	if tw.baseType != other.baseType {
		return false
	}

	if len(tw.parameters) != len(other.parameters) {
		return false
	}

	for i, param := range tw.parameters {
		if !param.Compare(other.parameters[i]) {
			return false
		}
	}

	return true
}

// Marshal converts the TypeWrapper to a string
func (tw *TypeWrapper) Marshal() string {
	if len(tw.parameters) == 0 {
		return tw.baseType
	}

	builder := strings.Builder{}
	builder.WriteString(tw.baseType)
	builder.WriteString("<")

	for i, param := range tw.parameters {
		builder.WriteString(param.Marshal())

		if i != len(tw.parameters)-1 {
			builder.WriteString(", ")
		}
	}

	builder.WriteString(">")

	return builder.String()
}
