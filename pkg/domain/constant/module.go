package constant

import "errors"

var ErrInvalidModuleName = errors.New("module name can only have alphanumeric characters and underscores with [a-zA-Z0-9_]")
