package constant

import "errors"

var (
	ErrInvalidVariableSetName = errors.New("variable set name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
	ErrEmptyVariableSetLabels = errors.New("variable set labels should not be empty")
)
