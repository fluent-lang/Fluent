package code

import "zyro/object"

// ZyroVariable represents a variable in a Zyro program
type ZyroVariable struct {
	constant bool
	value    object.ZyroObject
}

// NewZyroVariable creates a new Zyro variable
func NewZyroVariable(constant bool, value object.ZyroObject) ZyroVariable {
	return ZyroVariable{
		constant: constant,
		value:    value,
	}
}

// IsConstant returns true if the variable is a constant
func (sv *ZyroVariable) IsConstant() bool {
	return sv.constant
}

// GetValue returns the value of the variable
func (sv *ZyroVariable) GetValue() object.ZyroObject {
	return sv.value
}