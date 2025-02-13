package constant

import "errors"

var (
	ErrInvalidVariableName = errors.New("variable name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
	ErrInvalidVariableType = errors.New("invalid variable type, only PlainText and CipherText supported")
	ErrEmptyVariableSet    = errors.New("variable set should not be empty")
)
